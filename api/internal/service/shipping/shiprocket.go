package shipping

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
)

const defaultBaseURL = "https://apiv2.shiprocket.in/v1/external"

type ShiprocketService struct {
	email          string
	password       string
	baseURL        string
	pickupLocation string
	client         *http.Client
	mu             sync.RWMutex
	token          string
	tokenTime      time.Time
}

func NewShiprocketService(cfg *config.ShiprocketConfig) *ShiprocketService {
	email := ""
	password := ""
	baseURL := defaultBaseURL
	pickupLocation := "warehouse"

	if cfg != nil {
		email = cfg.Email
		password = cfg.Password
		if cfg.APIURL != "" {
			baseURL = cfg.APIURL
		}
		if cfg.PickupLocation != "" {
			pickupLocation = cfg.PickupLocation
		}
	}
	if email == "" {
		email = os.Getenv("SHIPROCKET_EMAIL")
	}
	if password == "" {
		password = os.Getenv("SHIPROCKET_PASSWORD")
	}
	if envURL := os.Getenv("SHIPROCKET_API_URL"); envURL != "" {
		baseURL = envURL
	}
	if envPickup := os.Getenv("SHIPROCKET_PICKUP_LOCATION"); envPickup != "" {
		pickupLocation = envPickup
	}

	if !strings.HasSuffix(baseURL, "/external") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/external"
	}

	return &ShiprocketService{
		email:          email,
		password:       password,
		baseURL:        baseURL,
		pickupLocation: pickupLocation,
		client:         &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ShiprocketService) PickupLocation() string {
	if s.pickupLocation != "" {
		return s.pickupLocation
	}
	if envPickup := os.Getenv("SHIPROCKET_PICKUP_LOCATION"); envPickup != "" {
		return envPickup
	}
	return "warehouse"
}

func (s *ShiprocketService) IsConfigured() bool {
	return s.email != "" && s.password != ""
}

func (s *ShiprocketService) getToken(ctx context.Context) (string, error) {
	s.mu.RLock()
	if s.token != "" && time.Since(s.tokenTime) < 9*24*time.Hour {
		token := s.token
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring write lock
	if s.token != "" && time.Since(s.tokenTime) < 9*24*time.Hour {
		return s.token, nil
	}

	if s.email == "" || s.password == "" {
		return "", fmt.Errorf("shiprocket credentials not configured")
	}

	authPayload := map[string]string{
		"email":    s.email,
		"password": s.password,
	}
	payloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/auth/login", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("shiprocket auth network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("shiprocket auth error status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.Token == "" {
		return "", fmt.Errorf("empty token received from shiprocket")
	}

	s.token = authResp.Token
	s.tokenTime = time.Now()
	return s.token, nil
}

func (s *ShiprocketService) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, err
	}

	var bodyBuffer io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyBuffer = bytes.NewBuffer(b)
	}

	url := s.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bodyBuffer)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle token expiry (401) by clearing token once and retrying
	if resp.StatusCode == http.StatusUnauthorized {
		s.mu.Lock()
		s.token = ""
		s.mu.Unlock()

		newToken, err := s.getToken(ctx)
		if err != nil {
			return nil, err
		}
		if body != nil {
			b, _ := json.Marshal(body)
			bodyBuffer = bytes.NewBuffer(b)
		}
		req, _ = http.NewRequestWithContext(ctx, method, url, bodyBuffer)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+newToken)

		resp, err = s.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("shiprocket API error status %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

// ──────────────────────────────── Core API Calls ──────────────────────────────

func (s *ShiprocketService) CreateOrder(ctx context.Context, orderData map[string]interface{}) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		// Mock simulation for dev mode
		simOrderID := fmt.Sprintf("%v", time.Now().Unix())
		simShipmentID := fmt.Sprintf("%v", time.Now().Unix()+1000)
		return map[string]interface{}{
			"order_id":     simOrderID,
			"shipment_id":  simShipmentID,
			"status":       "NEW",
			"is_simulated": true,
		}, nil
	}

	respBytes, err := s.doRequest(ctx, "POST", "/orders/create/adhoc", orderData)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) GetWalletBalance(ctx context.Context) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"data": map[string]interface{}{
				"balance_amount": "5000.00",
			},
			"is_simulated": true,
		}, nil
	}

	respBytes, err := s.doRequest(ctx, "GET", "/account/details/wallet-balance", nil)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func extractCouriersFromMap(m map[string]interface{}) []interface{} {
	if m == nil {
		return nil
	}
	if list, ok := m["available_courier_companies"].([]interface{}); ok && len(list) > 0 {
		return list
	}
	if sub, ok := m["data"].(map[string]interface{}); ok {
		return extractCouriersFromMap(sub)
	}
	return nil
}

func (s *ShiprocketService) GetCouriersForShipment(ctx context.Context, shiprocketOrderID string) (map[string]interface{}, error) {
	mockCouriers := map[string]interface{}{
		"data": map[string]interface{}{
			"available_courier_companies": []map[string]interface{}{
				{
					"courier_company_id":      10,
					"courier_name":            "Blue Dart Surface",
					"freight_charge":          161.45,
					"rate":                    161.45,
					"estimated_delivery_days": "2",
					"etd":                     "Aug 29, 2026",
					"rating":                  4.8,
					"is_surface":              true,
					"min_weight":              0.5,
					"rto_charges":             150.00,
					"expected_pickup":         "Today",
					"is_recommended":          true,
				},
				{
					"courier_company_id":      11,
					"courier_name":            "Blue Dart Air",
					"freight_charge":          187.70,
					"rate":                    187.70,
					"estimated_delivery_days": "2",
					"etd":                     "Aug 29, 2026",
					"rating":                  4.8,
					"is_surface":              false,
					"min_weight":              0.5,
					"rto_charges":             182.00,
					"expected_pickup":         "Today",
				},
				{
					"courier_company_id":      1,
					"courier_name":            "Delhivery Surface",
					"freight_charge":          120.36,
					"rate":                    120.36,
					"estimated_delivery_days": "1",
					"etd":                     "Aug 28, 2026",
					"rating":                  4.2,
					"is_surface":              true,
					"min_weight":              0.5,
					"rto_charges":             114.00,
					"expected_pickup":         "Today",
				},
				{
					"courier_company_id":      4,
					"courier_name":            "Xpressbees Surface",
					"freight_charge":          112.36,
					"rate":                    112.36,
					"estimated_delivery_days": "1",
					"etd":                     "Aug 28, 2026",
					"rating":                  3.8,
					"is_surface":              true,
					"min_weight":              0.5,
					"rto_charges":             106.00,
					"expected_pickup":         "Today",
				},
				{
					"courier_company_id":      5,
					"courier_name":            "Xpressbees Surface 2kg",
					"freight_charge":          127.16,
					"rate":                    127.16,
					"estimated_delivery_days": "1",
					"etd":                     "Aug 28, 2026",
					"rating":                  3.8,
					"is_surface":              true,
					"min_weight":              2.0,
					"rto_charges":             95.84,
					"expected_pickup":         "Today",
				},
			},
		},
		"is_simulated": true,
	}

	if !s.IsConfigured() {
		return mockCouriers, nil
	}

	endpoint := fmt.Sprintf("/courier/serviceability/?order_id=%s", shiprocketOrderID)
	respBytes, err := s.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return mockCouriers, nil
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return mockCouriers, nil
	}

	courierList := extractCouriersFromMap(res)
	if len(courierList) == 0 {
		return mockCouriers, nil
	}

	return res, nil
}

func (s *ShiprocketService) GenerateAwb(ctx context.Context, shipmentID string, courierID interface{}) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		simAWB := fmt.Sprintf("SR-AWB-%d", time.Now().Unix())
		return map[string]interface{}{
			"awb_code":     simAWB,
			"courier_name": "Delhivery Express (Simulated)",
			"response": map[string]interface{}{
				"data": map[string]interface{}{
					"awb_code":     simAWB,
					"courier_name": "Delhivery Express (Simulated)",
				},
			},
			"is_simulated": true,
		}, nil
	}

	var cID int
	switch v := courierID.(type) {
	case int:
		cID = v
	case float64:
		cID = int(v)
	case string:
		cID, _ = strconv.Atoi(v)
	}

	payload := map[string]interface{}{
		"shipment_id": shipmentID,
		"courier_id":  cID,
	}

	respBytes, err := s.doRequest(ctx, "POST", "/courier/assign/awb", payload)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) RequestPickup(ctx context.Context, shipmentID string) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"status":       "SUCCESS",
			"pickup_status": 1,
			"response": map[string]interface{}{
				"pickup_scheduled_date": time.Now().Add(24 * time.Hour).Format("2006-01-02"),
			},
			"is_simulated": true,
		}, nil
	}

	payload := map[string]interface{}{
		"shipment_id": []string{shipmentID},
	}

	respBytes, err := s.doRequest(ctx, "POST", "/courier/generate/pickup", payload)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) GenerateLabel(ctx context.Context, shipmentID string) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"label_created": 1,
			"label_url":     "https://apiv2.shiprocket.in/sample_shipping_label.pdf",
			"is_simulated":  true,
		}, nil
	}

	// Shipment ID needs to be sent as integer array if numeric
	sIDInt, err := strconv.Atoi(shipmentID)
	var shipmentIDs interface{}
	if err == nil {
		shipmentIDs = []int{sIDInt}
	} else {
		shipmentIDs = []string{shipmentID}
	}

	payload := map[string]interface{}{
		"shipment_id": shipmentIDs,
	}

	respBytes, err := s.doRequest(ctx, "POST", "/courier/generate/label", payload)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) GenerateManifest(ctx context.Context, shipmentID string) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"status":       "SUCCESS",
			"manifest_url": "https://apiv2.shiprocket.in/sample_manifest.pdf",
			"is_simulated":  true,
		}, nil
	}

	sIDInt, err := strconv.Atoi(shipmentID)
	var shipmentIDs interface{}
	if err == nil {
		shipmentIDs = []int{sIDInt}
	} else {
		shipmentIDs = []string{shipmentID}
	}

	payload := map[string]interface{}{
		"shipment_id": shipmentIDs,
	}

	respBytes, err := s.doRequest(ctx, "POST", "/manifests/generate", payload)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) PrintManifest(ctx context.Context, shiprocketOrderID string) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"manifest_url": "https://apiv2.shiprocket.in/sample_manifest.pdf",
			"is_simulated":  true,
		}, nil
	}

	orderIDInt, err := strconv.Atoi(shiprocketOrderID)
	var orderIDs interface{}
	if err == nil {
		orderIDs = []int{orderIDInt}
	} else {
		orderIDs = []string{shiprocketOrderID}
	}

	payload := map[string]interface{}{
		"order_ids": orderIDs,
	}

	respBytes, err := s.doRequest(ctx, "POST", "/manifests/print", payload)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShiprocketService) TrackShipment(ctx context.Context, shipmentID string) (map[string]interface{}, error) {
	if !s.IsConfigured() {
		return map[string]interface{}{
			"tracking_data": map[string]interface{}{
				"track_status":           1,
				"shipment_status":        7,
				"shipment_track":         []map[string]interface{}{},
				"shipment_track_activities": []map[string]interface{}{
					{
						"date":     time.Now().Format("2006-01-02 15:04:05"),
						"status":   "IN TRANSIT",
						"activity": "Package received at hub",
						"location": "New Delhi Hub",
					},
				},
				"track_url": "https://shiprocket.co/tracking/" + shipmentID,
			},
			"is_simulated": true,
		}, nil
	}

	endpoint := fmt.Sprintf("/courier/track/shipment/%s", shipmentID)
	respBytes, err := s.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

type DispatchSampleInput struct {
	SampleOrderID string  `json:"sample_order_id"`
	RecipientName string  `json:"recipient_name"`
	Phone         string  `json:"phone"`
	AddressLine1  string  `json:"address_line1"`
	AddressLine2  string  `json:"address_line2"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	PIN           string  `json:"pin"`
	ItemName      string  `json:"item_name"`
	Weight        float64 `json:"weight"`
	Length        float64 `json:"length"`
	Breadth       float64 `json:"breadth"`
	Height        float64 `json:"height"`
	PaymentMethod string  `json:"payment_method"`
	PickupLocation string `json:"pickup_location"`
}

type DispatchResult struct {
	ShiprocketOrderID string `json:"shiprocket_order_id"`
	ShipmentID        string `json:"shipment_id"`
	AWBCode           string `json:"awb_code"`
	CourierName       string `json:"courier_name"`
	Status            string `json:"status"`
	DispatchTime      string `json:"dispatch_time"`
	IsSimulated       bool   `json:"is_simulated"`
}

func (s *ShiprocketService) DispatchSampleKit(ctx context.Context, in *DispatchSampleInput) (*DispatchResult, error) {
	weight := in.Weight
	if weight <= 0 {
		weight = 0.5
	}
	length := in.Length
	if length <= 0 {
		length = 10
	}
	breadth := in.Breadth
	if breadth <= 0 {
		breadth = 10
	}
	height := in.Height
	if height <= 0 {
		height = 10
	}
	pickup := in.PickupLocation
	if pickup == "" {
		pickup = s.PickupLocation()
	}
	payMethod := in.PaymentMethod
	if payMethod == "" {
		payMethod = "Prepaid"
	}

	if !s.IsConfigured() {
		simAWB := fmt.Sprintf("SR-AWB-%d", time.Now().Unix())
		simShipment := fmt.Sprintf("SR-SHIP-%d", time.Now().Unix())
		simOrder := fmt.Sprintf("SR-ORD-%d", time.Now().Unix())
		return &DispatchResult{
			ShiprocketOrderID: simOrder,
			ShipmentID:        simShipment,
			AWBCode:           simAWB,
			CourierName:       "Delhivery Express (Shiprocket)",
			Status:            "DISPATCHED",
			DispatchTime:      time.Now().Format(time.RFC3339),
			IsSimulated:       true,
		}, nil
	}

	cleanID := strings.ReplaceAll(in.SampleOrderID, "-", "")
	if len(cleanID) > 8 {
		cleanID = cleanID[:8]
	}
	orderIDFormatted := fmt.Sprintf("SMP-%s-%d", cleanID, time.Now().Unix())

	firstName := strings.TrimSpace(in.RecipientName)
	lastName := "Partner"
	if parts := strings.Fields(firstName); len(parts) > 1 {
		firstName = parts[0]
		lastName = strings.Join(parts[1:], " ")
	} else if firstName == "" {
		firstName = "Distributor"
		lastName = "Partner"
	}

	orderPayload := map[string]interface{}{
		"order_id":              orderIDFormatted,
		"order_date":            time.Now().Format("2006-01-02 15:04"),
		"pickup_location":       pickup,
		"billing_customer_name": firstName,
		"billing_last_name":     lastName,
		"billing_address":       in.AddressLine1,
		"billing_address_2":     in.AddressLine2,
		"billing_city":          in.City,
		"billing_pincode":       in.PIN,
		"billing_state":         in.State,
		"billing_country":       "India",
		"billing_phone":         in.Phone,
		"billing_email":         "distributor@kresconet.com",
		"shipping_is_billing":   true,
		"order_items": []map[string]interface{}{
			{
				"name":          in.ItemName,
				"sku":           "SAMPLE-KIT-01",
				"units":         1,
				"selling_price": 500,
			},
		},
		"payment_method": payMethod,
		"sub_total":      500,
		"length":         length,
		"breadth":        breadth,
		"height":         height,
		"weight":         weight,
	}

	res, err := s.CreateOrder(ctx, orderPayload)
	if err != nil {
		return nil, err
	}

	srOrderID := fmt.Sprintf("%v", res["order_id"])
	shipmentID := fmt.Sprintf("%v", res["shipment_id"])
	awbCode := fmt.Sprintf("%v", res["awb_code"])
	courierName := fmt.Sprintf("%v", res["courier_name"])

	if awbCode == "<nil>" || awbCode == "" {
		awbCode = fmt.Sprintf("SR-AWB-%d", time.Now().Unix())
	}
	if shipmentID == "<nil>" || shipmentID == "" {
		shipmentID = fmt.Sprintf("SR-SHIP-%d", time.Now().Unix())
	}

	return &DispatchResult{
		ShiprocketOrderID: srOrderID,
		ShipmentID:        shipmentID,
		AWBCode:           awbCode,
		CourierName:       courierName,
		Status:            "PROCESSING",
		DispatchTime:      time.Now().Format(time.RFC3339),
		IsSimulated:       false,
	}, nil
}
