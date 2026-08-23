package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svcorder "github.com/arryaanjain/DistributorApprovalSystem/internal/service/order"
	"github.com/go-chi/chi/v5"
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

// GET /api/v1/catalogue/samples
func (h *CatalogueHandler) ListSamples(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListSampleCatalogue(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, products)
}

// GET /api/v1/admin/products
func (h *CatalogueHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListAllProductsAdmin(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, products)
}

// POST /api/v1/admin/products
func (h *CatalogueHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p repository.ProductRecord
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	id, err := h.svc.CreateProduct(r.Context(), &p)
	if err != nil {
		writeAppError(w, err)
		return
	}

	p.ID = id
	response.JSON(w, p)
}

// PUT /api/v1/admin/products/{id}
func (h *CatalogueHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "product ID required")
		return
	}

	var p repository.ProductRecord
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	p.ID = id

	if err := h.svc.UpdateProduct(r.Context(), &p); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "product updated successfully"})
}

// Stubs for legacy compatibility
func (h *CatalogueHandler) Get(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
func (h *CatalogueHandler) Delete(w http.ResponseWriter, r *http.Request) { stub(w, r) }
