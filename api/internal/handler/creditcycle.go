package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svccycle "github.com/arryaanjain/DistributorApprovalSystem/internal/service/ordercycle"
	svcrzp "github.com/arryaanjain/DistributorApprovalSystem/internal/service/razorpay"
	"github.com/go-chi/chi/v5"
)

type CreditCycleHandler struct {
	svc       *svccycle.Service
	rzpClient *svcrzp.Client
}

func NewCreditCycleHandler(svc *svccycle.Service, rzpClient *svcrzp.Client) *CreditCycleHandler {
	return &CreditCycleHandler{
		svc:       svc,
		rzpClient: rzpClient,
	}
}

// InitCycle: POST /api/v1/credit-cycles/init
func (h *CreditCycleHandler) InitCycle(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var input svccycle.CreateCycleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.svc.CreateCycle(r.Context(), distID, input)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// CompleteESign: POST /api/v1/credit-cycles/orders/{order_id}/complete-esign
func (h *CreditCycleHandler) CompleteESign(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")

	res, err := h.svc.CompleteESign(r.Context(), orderID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// SetupMandate: POST /api/v1/credit-cycles/orders/{order_id}/setup-mandate
func (h *CreditCycleHandler) SetupMandate(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")

	res, err := h.svc.SetupMandate(r.Context(), orderID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// GetCycleDetails: GET /api/v1/credit-cycles/orders/{order_id}
func (h *CreditCycleHandler) GetCycleDetails(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	orderID := chi.URLParam(r, "order_id")

	res, err := h.svc.GetCycleDetails(r.Context(), distID, orderID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// GetMine: GET /api/v1/credit-cycles/me
func (h *CreditCycleHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	cycles, err := h.svc.GetDistributorCycles(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]interface{}{
		"cycles": cycles,
	})
}

// ESignCallback: GET /api/v1/credit-cycles/orders/{order_id}/esign-callback
func (h *CreditCycleHandler) ESignCallback(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")

	// Trigger completion & mandate creation
	_, _ = h.svc.CompleteESign(r.Context(), orderID)

	// Redirect back to distributor dashboard order cycle flow
	http.Redirect(w, r, "http://localhost:3000/orders?esign=success&order_id="+orderID, http.StatusFound)
}

// RazorpayWebhook: POST /api/v1/credit-cycles/webhooks/razorpay
func (h *CreditCycleHandler) RazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading payload", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Razorpay-Signature")
	if h.rzpClient != nil && sig != "" {
		if !h.rzpClient.VerifyWebhookSignature(body, sig) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	eventType, _ := payload["event"].(string)
	eventPayload, _ := payload["payload"].(map[string]interface{})

	if eventType != "" && eventPayload != nil {
		_ = h.svc.ProcessMandateWebhook(r.Context(), eventType, eventPayload)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
