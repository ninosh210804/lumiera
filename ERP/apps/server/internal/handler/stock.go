package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type stockHandler struct {
	stock *service.StockService
}

// POST /api/v1/stock/write-off
type writeOffRequest struct {
	IngredientID string  `json:"ingredient_id"`
	Qty          float64 `json:"qty"`
	Reason       string  `json:"reason"`
	Note         string  `json:"note"`
}

func (h *stockHandler) writeOff(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req writeOffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	ingID, err := uuid.Parse(req.IngredientID)
	if err != nil {
		mw.Error(w, badRequestf("invalid ingredient_id"))
		return
	}
	if req.Qty <= 0 {
		mw.Error(w, badRequestf("qty must be > 0"))
		return
	}
	if req.Reason == "" {
		mw.Error(w, badRequestf("reason is required"))
		return
	}

	mv, err := h.stock.WriteOff(r.Context(), service.WriteOffInput{
		LocationID:   locID,
		IngredientID: ingID,
		Qty:          req.Qty,
		Reason:       req.Reason,
		Note:         req.Note,
		CreatedBy:    claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, mv)
}

// POST /api/v1/stock/receive
type receiveStockRequest struct {
	SupplierID string                   `json:"supplier_id"`
	Notes      string                   `json:"notes"`
	Items      []receiveStockItemRequest `json:"items"`
}

type receiveStockItemRequest struct {
	IngredientID string  `json:"ingredient_id"`
	Qty          float64 `json:"qty"`
	UnitCost     float64 `json:"unit_cost"`
	ExpiresAt    string  `json:"expires_at"` // optional YYYY-MM-DD
}

func (h *stockHandler) receiveStock(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req receiveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if len(req.Items) == 0 {
		mw.Error(w, badRequestf("items is required"))
		return
	}

	in := service.ReceiveStockInput{
		LocationID: locID,
		Notes:      req.Notes,
		CreatedBy:  claims.UserID,
	}
	if req.SupplierID != "" {
		if id, err := uuid.Parse(req.SupplierID); err == nil {
			in.SupplierID = &id
		}
	}

	for _, item := range req.Items {
		ingID, err := uuid.Parse(item.IngredientID)
		if err != nil {
			mw.Error(w, badRequestf("invalid ingredient_id: "+item.IngredientID))
			return
		}
		if item.Qty <= 0 {
			mw.Error(w, badRequestf("item qty must be > 0"))
			return
		}
		if item.UnitCost < 0 {
			mw.Error(w, badRequestf("unit_cost must be >= 0"))
			return
		}
		ri := service.ReceiveStockItemInput{
			IngredientID: ingID,
			Qty:          item.Qty,
			UnitCost:     item.UnitCost,
		}
		if item.ExpiresAt != "" {
			if t, err := time.Parse("2006-01-02", item.ExpiresAt); err == nil {
				ri.ExpiresAt = &t
			}
		}
		in.Items = append(in.Items, ri)
	}

	receipt, err := h.stock.ReceiveStock(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, receipt)
}

// POST /api/v1/stock/count
type createCountRequest struct {
	Items []countItemRequest `json:"items"`
}

type countItemRequest struct {
	IngredientID string  `json:"ingredient_id"`
	ActualQty    float64 `json:"actual_qty"`
}

func (h *stockHandler) createCount(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req createCountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if len(req.Items) == 0 {
		mw.Error(w, badRequestf("items is required"))
		return
	}

	in := service.CreateCountInput{
		LocationID:  locID,
		PerformedBy: claims.UserID,
	}
	for _, item := range req.Items {
		ingID, err := uuid.Parse(item.IngredientID)
		if err != nil {
			mw.Error(w, badRequestf("invalid ingredient_id: "+item.IngredientID))
			return
		}
		if item.ActualQty < 0 {
			mw.Error(w, badRequestf("actual_qty must be >= 0"))
			return
		}
		in.Items = append(in.Items, service.CountItemInput{
			IngredientID: ingID,
			ActualQty:    item.ActualQty,
		})
	}

	count, err := h.stock.CreateCount(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, count)
}

// GET /api/v1/stock/movements?ingredient_id=&from=&to=
func (h *stockHandler) listMovements(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	q := r.URL.Query()
	var ingID *uuid.UUID
	if s := q.Get("ingredient_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			ingID = &id
		}
	}

	from := time.Now().Truncate(24 * time.Hour)
	to := from.Add(24 * time.Hour)
	if f := q.Get("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			from = t
		}
	}
	if t := q.Get("to"); t != "" {
		if tp, err := time.Parse(time.RFC3339, t); err == nil {
			to = tp
		}
	}

	movements, err := h.stock.ListMovements(r.Context(), locID, ingID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, movements)
}

// GET /api/v1/stock/counts
func (h *stockHandler) listCounts(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	counts, err := h.stock.ListCounts(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, counts)
}

// ─── Suppliers ────────────────────────────────────────────────────────────────

type suppliersHandler struct {
	stock *service.StockService
}

// GET /api/v1/suppliers
func (h *suppliersHandler) list(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.stock.ListSuppliers(r.Context())
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, suppliers)
}

// POST /api/v1/suppliers
type createSupplierRequest struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	BIN     string `json:"bin"`
	Address string `json:"address"`
}

func (h *suppliersHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	s, err := h.stock.CreateSupplier(r.Context(), service.CreateSupplierInput{
		Name:    req.Name,
		Contact: req.Contact,
		Phone:   req.Phone,
		BIN:     req.BIN,
		Address: req.Address,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, s)
}
