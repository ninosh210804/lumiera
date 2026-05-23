package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type inventoryHandler struct {
	ing *service.IngredientService
}

// GET /api/v1/ingredients?location_id=&low_stock=true
func (h *inventoryHandler) listIngredients(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	var items interface{}
	if r.URL.Query().Get("low_stock") == "true" {
		items, err = h.ing.ListLowStock(r.Context(), locID)
	} else {
		items, err = h.ing.List(r.Context(), locID)
	}
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, items)
}

// GET /api/v1/ingredients/{id}
func (h *inventoryHandler) getIngredient(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid ingredient id"))
		return
	}
	item, err := h.ing.Get(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, item)
}

// POST /api/v1/ingredients
type createIngredientRequest struct {
	Name                 string  `json:"name"`
	Unit                 string  `json:"unit"`
	IsPerishable         bool    `json:"is_perishable"`
	DefaultShelfLifeDays *int32  `json:"default_shelf_life_days"`
	MinStockAlert        float64 `json:"min_stock_alert"`
}

func (h *inventoryHandler) createIngredient(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}
	var req createIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	validUnits := map[string]bool{"g": true, "kg": true, "ml": true, "l": true, "pcs": true}
	if !validUnits[req.Unit] {
		mw.Error(w, badRequestf("unit must be one of: g, kg, ml, l, pcs"))
		return
	}
	item, err := h.ing.Create(r.Context(), service.CreateIngredientInput{
		LocationID:           locID,
		Name:                 req.Name,
		Unit:                 req.Unit,
		IsPerishable:         req.IsPerishable,
		DefaultShelfLifeDays: req.DefaultShelfLifeDays,
		MinStockAlert:        req.MinStockAlert,
		CreatedBy:            claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, item)
}

// PUT /api/v1/ingredients/{id}
type updateIngredientRequest struct {
	Name                 string  `json:"name"`
	Unit                 string  `json:"unit"`
	IsPerishable         bool    `json:"is_perishable"`
	DefaultShelfLifeDays *int32  `json:"default_shelf_life_days"`
	MinStockAlert        float64 `json:"min_stock_alert"`
	IsActive             bool    `json:"is_active"`
}

func (h *inventoryHandler) updateIngredient(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid ingredient id"))
		return
	}
	var req updateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	item, err := h.ing.Update(r.Context(), service.UpdateIngredientInput{
		ID:                   id,
		Name:                 req.Name,
		Unit:                 req.Unit,
		IsPerishable:         req.IsPerishable,
		DefaultShelfLifeDays: req.DefaultShelfLifeDays,
		MinStockAlert:        req.MinStockAlert,
		IsActive:             req.IsActive,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, item)
}

// DELETE /api/v1/ingredients/{id}
func (h *inventoryHandler) deleteIngredient(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid ingredient id"))
		return
	}
	if err := h.ing.Delete(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
