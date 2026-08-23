package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcorder "github.com/arryaanjain/DistributorApprovalSystem/internal/service/order"
)

type CatalogueHandler struct {
	svc *svcorder.Service
}

func NewCatalogueHandler(svc *svcorder.Service) *CatalogueHandler {
	return &CatalogueHandler{svc: svc}
}

// GET /api/v1/catalogue
func (h *CatalogueHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListCatalogue(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, products)
}

// Stubs for legacy registry methods
func (h *CatalogueHandler) Get(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
func (h *CatalogueHandler) Create(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *CatalogueHandler) Update(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *CatalogueHandler) Delete(w http.ResponseWriter, r *http.Request) { stub(w, r) }
