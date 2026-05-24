package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type salesHandler struct {
	sales *service.SaleService
}

func (h *salesHandler) list(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	events, err := h.sales.List(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, events)
}

func (h *salesHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid sale event id"))
		return
	}
	ev, err := h.sales.Get(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, ev)
}

type createSaleRequest struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (h *salesHandler) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	var req createSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	ev, err := h.sales.Create(r.Context(), claims.LocationID, claims.UserID, req.Name, req.IsActive)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, ev)
}

type setActiveRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *salesHandler) setActive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid sale event id"))
		return
	}
	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if err := h.sales.SetActive(r.Context(), id, req.IsActive); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *salesHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid sale event id"))
		return
	}
	if err := h.sales.Delete(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type saleItemRequest struct {
	ProductID string  `json:"product_id"`
	SalePrice float64 `json:"sale_price"`
}

func (h *salesHandler) addItem(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid sale event id"))
		return
	}
	var req saleItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		mw.Error(w, badRequestf("invalid product_id"))
		return
	}
	if err := h.sales.AddItem(r.Context(), eventID, productID, req.SalePrice); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *salesHandler) removeItem(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid sale event id"))
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	if err := h.sales.RemoveItem(r.Context(), eventID, productID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
