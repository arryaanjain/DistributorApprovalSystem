package handler

import (
	"net/http"
	"strconv"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	"github.com/go-chi/chi/v5"
)

// DistributorHandler handles distributor profile endpoints.
type DistributorHandler struct {
	repo *repository.DistributorRepository
}

func NewDistributorHandler(repo *repository.DistributorRepository) *DistributorHandler {
	return &DistributorHandler{repo: repo}
}

// GET /api/v1/distributors/me
func (h *DistributorHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	dist, err := h.repo.GetByID(r.Context(), distID)
	if err != nil || dist == nil {
		response.NotFound(w, "distributor not found")
		return
	}
	response.JSON(w, dist)
}

// GET /api/v1/distributors — employee only
func (h *DistributorHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 || limit > 100 {
		limit = 25
	}

	list, total, err := h.repo.ListAll(r.Context(), limit, offset)
	if err != nil {
		response.InternalError(w)
		return
	}

	pages := total / limit
	if total%limit != 0 {
		pages++
	}

	response.WithMeta(w, list, &response.Meta{
		Page:       offset/limit + 1,
		PerPage:    limit,
		Total:      total,
		TotalPages: pages,
	})
}

// GET /api/v1/distributors/{id}
func (h *DistributorHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dist, err := h.repo.GetByID(r.Context(), id)
	if err != nil || dist == nil {
		response.NotFound(w, "distributor not found")
		return
	}
	response.JSON(w, dist)
}

// GET /api/v1/distributors/{id}/summary — returns distributor + business profile + documents
func (h *DistributorHandler) Summary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	dist, err := h.repo.GetByID(r.Context(), id)
	if err != nil || dist == nil {
		response.NotFound(w, "distributor not found")
		return
	}

	bp, _ := h.repo.GetBusinessProfile(r.Context(), id)
	docs, _ := h.repo.GetBusinessDocuments(r.Context(), id)
	app, _ := h.repo.GetActiveApplication(r.Context(), id)

	response.JSON(w, map[string]interface{}{
		"distributor":       dist,
		"business_profile":  bp,
		"documents":         docs,
		"active_application": app,
	})
}
