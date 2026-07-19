package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type onlineOrdersHandler struct {
	online *service.OnlineOrderService
}

// ─── Public (customer delivery app) ─────────────────────────────────────────

type placeOnlineOrderRequest struct {
	LocationID     string             `json:"location_id"`
	CustomerPhone  string             `json:"customer_phone"`
	DeliveryOffice string             `json:"delivery_office"`
	DeliveryNote   string             `json:"delivery_note"`
	Items          []orderItemRequest `json:"items"`
}

// POST /api/v1/clients/orders — place a delivery order (no auth).
func (h *onlineOrdersHandler) place(w http.ResponseWriter, r *http.Request) {
	var req placeOnlineOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	locID, err := uuid.Parse(req.LocationID)
	if err != nil {
		mw.Error(w, badRequestf("invalid location_id"))
		return
	}
	if len(req.Items) == 0 {
		mw.Error(w, badRequestf("items is required"))
		return
	}
	in := service.PlaceOnlineOrderInput{
		LocationID:     locID,
		CustomerPhone:  req.CustomerPhone,
		DeliveryOffice: req.DeliveryOffice,
		DeliveryNote:   req.DeliveryNote,
	}
	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			mw.Error(w, badRequestf("invalid product_id: "+item.ProductID))
			return
		}
		if item.Qty <= 0 {
			mw.Error(w, badRequestf("item qty must be > 0"))
			return
		}
		in.Items = append(in.Items, service.OrderItemInput{ProductID: productID, Qty: item.Qty})
	}
	order, err := h.online.PlaceOrder(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, order)
}

// GET /api/v1/clients/orders/{id} — poll a delivery order's status (no auth).
func (h *onlineOrdersHandler) track(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	order, err := h.online.Track(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, order)
}

// ─── Staff (barista queue) ──────────────────────────────────────────────────

// GET /api/v1/online-orders — active queue for the caller's location.
func (h *onlineOrdersHandler) list(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}
	orders, err := h.online.ListActive(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, orders)
}

type setOnlineStatusRequest struct {
	Status string `json:"status"`
}

// POST /api/v1/online-orders/{id}/status — advance the fulfilment status.
func (h *onlineOrdersHandler) setStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	var req setOnlineStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Status == "" {
		mw.Error(w, badRequestf("status is required"))
		return
	}
	claims, _ := mw.ClaimsFrom(r.Context())
	order, err := h.online.Advance(r.Context(), id, req.Status, claims.UserID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, order)
}

type completeOnlineOrderRequest struct {
	PaymentMethod string `json:"payment_method"`
	ShiftID       string `json:"shift_id"`
}

// POST /api/v1/online-orders/{id}/complete — collect payment and book the order.
func (h *onlineOrdersHandler) complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	var req completeOnlineOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.PaymentMethod == "" {
		mw.Error(w, badRequestf("payment_method is required"))
		return
	}
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}
	in := service.CompleteOnlineOrderInput{
		OnlineOrderID: id,
		BaristaID:     claims.UserID,
		LocationID:    locID,
		PaymentMethod: req.PaymentMethod,
	}
	if req.ShiftID != "" {
		if sid, err := uuid.Parse(req.ShiftID); err == nil {
			in.ShiftID = &sid
		}
	}
	order, err := h.online.Complete(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, order)
}
