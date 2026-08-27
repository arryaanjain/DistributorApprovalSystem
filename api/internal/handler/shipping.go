package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/service/shipping"
)

type ShippingHandler struct {
	srSvc     *shipping.ShiprocketService
	orderRepo *repository.OrderRepository
}

func NewShippingHandler(srSvc *shipping.ShiprocketService, orderRepo *repository.OrderRepository) *ShippingHandler {
	return &ShippingHandler{
		srSvc:     srSvc,
		orderRepo: orderRepo,
	}
}

// ──────────────────────────────── 1. Create Shipment ─────────────────────────

type createShipmentReq struct {
	Weight         float64 `json:"weight"`
	Length         float64 `json:"length"`
	Breadth        float64 `json:"breadth"`
	Height         float64 `json:"height"`
	PaymentMethod  string  `json:"payment_method"`
	PickupLocation string  `json:"pickup_location"`
}

func (h *ShippingHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		response.BadRequest(w, "order id is required")
		return
	}

	var req createShipmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if req.Weight <= 0 {
		response.BadRequest(w, "Weight must be greater than 0")
		return
	}
	if req.Length <= 0 || req.Breadth <= 0 || req.Height <= 0 {
		response.BadRequest(w, "All package dimensions must be greater than 0")
		return
	}
	if req.PickupLocation == "" {
		req.PickupLocation = h.srSvc.PickupLocation()
	}
	if req.PaymentMethod != "Prepaid" && req.PaymentMethod != "COD" {
		req.PaymentMethod = "Prepaid"
	}

	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("sample order not found"))
		return
	}

	recipientName := "Distributor Partner"
	if sampleOrder.DistributorName != nil && *sampleOrder.DistributorName != "" {
		recipientName = *sampleOrder.DistributorName
	}
	phone := "9999999999"
	if sampleOrder.DistributorMobile != nil && *sampleOrder.DistributorMobile != "" {
		phone = *sampleOrder.DistributorMobile
	}

	line1 := "Default Warehouse Delivery"
	line2 := ""
	city := "New Delhi"
	state := "Delhi"
	pin := "110001"

	if sampleOrder.ShippingAddress != nil {
		line1 = sampleOrder.ShippingAddress.AddressLine1
		if sampleOrder.ShippingAddress.AddressLine2 != nil {
			line2 = *sampleOrder.ShippingAddress.AddressLine2
		}
		city = sampleOrder.ShippingAddress.City
		state = sampleOrder.ShippingAddress.State
		pin = sampleOrder.ShippingAddress.PIN
		if sampleOrder.ShippingAddress.Phone != nil && *sampleOrder.ShippingAddress.Phone != "" {
			phone = *sampleOrder.ShippingAddress.Phone
		}
	}

	// Clean phone number
	phoneClean := strings.ReplaceAll(phone, "+91", "")
	phoneClean = strings.ReplaceAll(phoneClean, " ", "")

	in := &shipping.DispatchSampleInput{
		SampleOrderID:  sampleOrder.ID,
		RecipientName:  recipientName,
		Phone:          phoneClean,
		AddressLine1:   line1,
		AddressLine2:   line2,
		City:           city,
		State:          state,
		PIN:            pin,
		ItemName:       "Distributor Sample Kit",
		Weight:         req.Weight,
		Length:         req.Length,
		Breadth:        req.Breadth,
		Height:         req.Height,
		PaymentMethod:  req.PaymentMethod,
		PickupLocation: req.PickupLocation,
	}

	res, err := h.srSvc.DispatchSampleKit(r.Context(), in)
	if err != nil {
		writeAppError(w, apperrors.Internal("Shiprocket shipment creation failed: "+err.Error(), err))
		return
	}

	_ = h.orderRepo.UpdateSampleOrderShipment(r.Context(), sampleOrder.ID, res.ShiprocketOrderID, res.ShipmentID, req.Weight, req.Length, req.Breadth, req.Height)

	response.JSON(w, map[string]interface{}{
		"success":             true,
		"message":             "Shipment created successfully",
		"shiprocket_order_id": res.ShiprocketOrderID,
		"shipment_id":        res.ShipmentID,
		"is_simulated":        res.IsSimulated,
	})
}

// ──────────────────────────────── 2. Wallet Balance ───────────────────────────

func (h *ShippingHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	data, err := h.srSvc.GetWalletBalance(r.Context())
	if err != nil {
		writeAppError(w, apperrors.Internal("Failed to fetch wallet balance: "+err.Error(), err))
		return
	}
	response.JSON(w, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ──────────────────────────────── 3. Available Couriers ───────────────────────

func (h *ShippingHandler) GetAvailableCouriers(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	srOrderID := ""
	if sampleOrder.ShiprocketOrderID != nil {
		srOrderID = *sampleOrder.ShiprocketOrderID
	}
	if srOrderID == "" {
		response.BadRequest(w, "No Shiprocket order found. Create a shipment first.")
		return
	}

	data, err := h.srSvc.GetCouriersForShipment(r.Context(), srOrderID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Failed to fetch available couriers: "+err.Error(), err))
		return
	}

	response.JSON(w, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ──────────────────────────────── 4. Assign Courier (Generates AWB) ──────────

type assignCourierReq struct {
	CourierID   interface{} `json:"courier_id"`
	CourierRate float64     `json:"courier_rate"`
}

func (h *ShippingHandler) AssignCourier(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	shipmentID := ""
	if sampleOrder.ShipmentID != nil {
		shipmentID = *sampleOrder.ShipmentID
	}
	if shipmentID == "" {
		response.BadRequest(w, "No shipment found for this order. Create a shipment first.")
		return
	}

	var req assignCourierReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if req.CourierID == nil {
		response.BadRequest(w, "courier_id is required")
		return
	}

	// Pre-flight wallet balance check
	walletData, wErr := h.srSvc.GetWalletBalance(r.Context())
	if wErr == nil && walletData != nil {
		var balance float64
		if innerData, ok := walletData["data"].(map[string]interface{}); ok {
			if balStr, ok := innerData["balance_amount"].(string); ok {
				balance, _ = strconv.ParseFloat(balStr, 64)
			} else if balNum, ok := innerData["balance_amount"].(float64); ok {
				balance = balNum
			}
		}
		if req.CourierRate > 0 && balance < req.CourierRate {
			w.WriteHeader(http.StatusPaymentRequired)
			response.JSON(w, map[string]interface{}{
				"error":           fmt.Sprintf("Insufficient Shiprocket wallet balance! You need ₹%.2f but only have ₹%.2f. Please recharge your wallet.", req.CourierRate, balance),
				"wallet_balance":  balance,
				"required_amount": req.CourierRate,
			})
			return
		}
	}

	awbResp, err := h.srSvc.GenerateAwb(r.Context(), shipmentID, req.CourierID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Failed to assign courier & generate AWB: "+err.Error(), err))
		return
	}

	awbCode := ""
	courierName := ""
	if respData, ok := awbResp["response"].(map[string]interface{}); ok {
		if inner, ok := respData["data"].(map[string]interface{}); ok {
			if val, ok := inner["awb_code"].(string); ok {
				awbCode = val
			}
			if val, ok := inner["courier_name"].(string); ok {
				courierName = val
			}
		}
	}
	if awbCode == "" {
		if val, ok := awbResp["awb_code"].(string); ok {
			awbCode = val
		}
	}
	if courierName == "" {
		if val, ok := awbResp["courier_name"].(string); ok {
			courierName = val
		}
	}

	if awbCode == "" {
		awbCode = fmt.Sprintf("SR-AWB-%s", orderID[:8])
	}
	if courierName == "" {
		courierName = "Shiprocket Express Partner"
	}

	_ = h.orderRepo.UpdateSampleOrderAWB(r.Context(), sampleOrder.ID, awbCode, courierName)

	response.JSON(w, map[string]interface{}{
		"success":      true,
		"message":      "Courier assigned and AWB generated successfully",
		"awb_code":     awbCode,
		"courier_name": courierName,
		"response":     awbResp,
	})
}

// ──────────────────────────────── 5. Request Pickup ──────────────────────────

func (h *ShippingHandler) RequestPickup(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	shipmentID := ""
	if sampleOrder.ShipmentID != nil {
		shipmentID = *sampleOrder.ShipmentID
	}
	if shipmentID == "" {
		response.BadRequest(w, "No shipment found for this order.")
		return
	}

	pickupResp, err := h.srSvc.RequestPickup(r.Context(), shipmentID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Pickup request failed: "+err.Error(), err))
		return
	}

	_ = h.orderRepo.UpdateSampleOrderPickupStatus(r.Context(), sampleOrder.ID, "PICKUP_REQUESTED")

	response.JSON(w, map[string]interface{}{
		"success": true,
		"message": "Pickup requested successfully",
		"data":    pickupResp,
	})
}

// ──────────────────────────────── 6. Generate Label ──────────────────────────

func (h *ShippingHandler) GenerateLabel(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	shipmentID := ""
	if sampleOrder.ShipmentID != nil {
		shipmentID = *sampleOrder.ShipmentID
	}
	if shipmentID == "" {
		response.BadRequest(w, "No shipment found for this order.")
		return
	}

	labelResp, err := h.srSvc.GenerateLabel(r.Context(), shipmentID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Label generation failed: "+err.Error(), err))
		return
	}

	labelURL := ""
	if val, ok := labelResp["label_url"].(string); ok {
		labelURL = val
	}

	if labelURL != "" {
		_ = h.orderRepo.UpdateSampleOrderLabel(r.Context(), sampleOrder.ID, labelURL)
	}

	response.JSON(w, map[string]interface{}{
		"success":   true,
		"label_url": labelURL,
		"data":      labelResp,
	})
}

// ──────────────────────────────── 7. Generate Manifest ───────────────────────

func (h *ShippingHandler) GenerateManifest(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	shipmentID := ""
	if sampleOrder.ShipmentID != nil {
		shipmentID = *sampleOrder.ShipmentID
	}
	if shipmentID == "" {
		response.BadRequest(w, "No shipment found for this order.")
		return
	}

	manifestResp, err := h.srSvc.GenerateManifest(r.Context(), shipmentID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Manifest generation failed: "+err.Error(), err))
		return
	}

	manifestURL := ""
	if sampleOrder.ShiprocketOrderID != nil && *sampleOrder.ShiprocketOrderID != "" {
		printResp, _ := h.srSvc.PrintManifest(r.Context(), *sampleOrder.ShiprocketOrderID)
		if printResp != nil {
			if val, ok := printResp["manifest_url"].(string); ok {
				manifestURL = val
			}
		}
	}

	if manifestURL == "" {
		if val, ok := manifestResp["manifest_url"].(string); ok {
			manifestURL = val
		}
	}

	if manifestURL != "" {
		_ = h.orderRepo.UpdateSampleOrderManifest(r.Context(), sampleOrder.ID, manifestURL)
	}

	response.JSON(w, map[string]interface{}{
		"success":      true,
		"manifest_url": manifestURL,
		"data":         manifestResp,
	})
}

// ──────────────────────────────── 8. Track Shipment ──────────────────────────

func (h *ShippingHandler) TrackShipment(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), orderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("Sample order not found"))
		return
	}

	shipmentID := ""
	if sampleOrder.ShipmentID != nil {
		shipmentID = *sampleOrder.ShipmentID
	}
	if shipmentID == "" {
		response.BadRequest(w, "No shipment found for this order")
		return
	}

	tracking, err := h.srSvc.TrackShipment(r.Context(), shipmentID)
	if err != nil {
		writeAppError(w, apperrors.Internal("Failed to track shipment: "+err.Error(), err))
		return
	}

	response.JSON(w, tracking)
}

// ──────────────────────────────── 9. Webhook Listener ────────────────────────

func (h *ShippingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	awb := ""
	if val, ok := payload["awb"].(string); ok {
		awb = val
	} else if val, ok := payload["awb_code"].(string); ok {
		awb = val
	}

	if awb == "" {
		response.JSON(w, map[string]interface{}{"status": "ignored", "reason": "Missing AWB"})
		return
	}

	currentStatus := ""
	if val, ok := payload["current_status"].(string); ok {
		currentStatus = strings.ToUpper(strings.TrimSpace(val))
	} else if val, ok := payload["status"].(string); ok {
		currentStatus = strings.ToUpper(strings.TrimSpace(val))
	}

	sampleOrder, err := h.orderRepo.GetSampleOrderByAWB(r.Context(), awb)
	if err != nil || sampleOrder == nil {
		response.JSON(w, map[string]interface{}{"status": "order_not_found", "awb": awb})
		return
	}

	switch currentStatus {
	case "OUT FOR DELIVERY", "OUT_FOR_DELIVERY":
		_ = h.orderRepo.UpdateSampleOrderStatus(r.Context(), sampleOrder.ID, "OUT_FOR_DELIVERY")
	case "DELIVERED":
		_ = h.orderRepo.UpdateSampleOrderStatus(r.Context(), sampleOrder.ID, "DELIVERED")
	}

	response.JSON(w, map[string]interface{}{
		"status":         "success",
		"order_id":       sampleOrder.ID,
		"current_status": currentStatus,
	})
}

// ──────────────────────────────── Deprecated Legacy Handler ───────────────────

type dispatchReq struct {
	SampleOrderID string `json:"sample_order_id"`
}

func (h *ShippingHandler) DispatchSampleOrder(w http.ResponseWriter, r *http.Request) {
	var req dispatchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if req.SampleOrderID == "" {
		response.BadRequest(w, "sample_order_id is required")
		return
	}

	sampleOrder, err := h.orderRepo.GetSampleOrderByID(r.Context(), req.SampleOrderID)
	if err != nil || sampleOrder == nil {
		writeAppError(w, apperrors.NotFound("sample order not found"))
		return
	}

	name := "Distributor Partner"
	if sampleOrder.DistributorName != nil && *sampleOrder.DistributorName != "" {
		name = *sampleOrder.DistributorName
	}
	phone := "9999999999"
	if sampleOrder.DistributorMobile != nil && *sampleOrder.DistributorMobile != "" {
		phone = *sampleOrder.DistributorMobile
	}

	line1 := "Default Warehouse Delivery"
	line2 := ""
	city := "New Delhi"
	state := "Delhi"
	pin := "110001"

	if sampleOrder.ShippingAddress != nil {
		line1 = sampleOrder.ShippingAddress.AddressLine1
		if sampleOrder.ShippingAddress.AddressLine2 != nil {
			line2 = *sampleOrder.ShippingAddress.AddressLine2
		}
		city = sampleOrder.ShippingAddress.City
		state = sampleOrder.ShippingAddress.State
		pin = sampleOrder.ShippingAddress.PIN
		if sampleOrder.ShippingAddress.Phone != nil && *sampleOrder.ShippingAddress.Phone != "" {
			phone = *sampleOrder.ShippingAddress.Phone
		}
	}

	in := &shipping.DispatchSampleInput{
		SampleOrderID:  sampleOrder.ID,
		RecipientName:  name,
		Phone:          phone,
		AddressLine1:   line1,
		AddressLine2:   line2,
		City:           city,
		State:          state,
		PIN:            pin,
		ItemName:       "Sample Kit",
		Weight:         0.5,
		Length:         10,
		Breadth:        10,
		Height:         10,
		PaymentMethod:  "Prepaid",
		PickupLocation: h.srSvc.PickupLocation(),
	}

	result, err := h.srSvc.DispatchSampleKit(r.Context(), in)
	if err != nil {
		writeAppError(w, apperrors.Internal("shiprocket dispatch failed", err))
		return
	}

	_ = h.orderRepo.UpdateSampleOrderShipment(r.Context(), sampleOrder.ID, result.ShiprocketOrderID, result.ShipmentID, 0.5, 10, 10, 10)
	_ = h.orderRepo.UpdateSampleOrderAWB(r.Context(), sampleOrder.ID, result.AWBCode, result.CourierName)

	response.JSON(w, result)
}
