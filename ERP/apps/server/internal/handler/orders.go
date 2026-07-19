package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type ordersHandler struct {
	orders *service.OrderService
}

// GET /api/v1/orders/payment-methods
func (h *ordersHandler) listPaymentMethods(w http.ResponseWriter, r *http.Request) {
	methods, err := h.orders.ListPaymentMethods(r.Context())
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, methods)
}

// POST /api/v1/orders
type createOrderRequest struct {
	Items              []orderItemRequest `json:"items"`
	Payments           []paymentRequest   `json:"payments"`
	CustomerPhone      string             `json:"customer_phone"`
	LoyaltyPointsToUse float64            `json:"loyalty_points_to_use"`
	ManualDiscountPct  float64            `json:"manual_discount_pct"`
	ClientUUID         string             `json:"client_uuid"`
	ShiftID            string             `json:"shift_id"`
	Comp               bool               `json:"comp"`
	CompRecipient      string             `json:"comp_recipient"`
}

type orderItemRequest struct {
	ProductID         string   `json:"product_id"`
	Qty               float64  `json:"qty"`
	ModifierOptionIDs []string `json:"modifier_option_ids"`
	ClientUUID        string   `json:"client_uuid"`
}

type paymentRequest struct {
	MethodCode  string  `json:"method"`
	Amount      float64 `json:"amount"`
	ExternalRef string  `json:"external_ref"`
}

func (h *ordersHandler) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if len(req.Items) == 0 {
		mw.Error(w, badRequestf("items is required"))
		return
	}
	if len(req.Payments) == 0 && !req.Comp {
		mw.Error(w, badRequestf("payments is required"))
		return
	}
	if req.Comp && req.CompRecipient == "" {
		mw.Error(w, badRequestf("comp_recipient is required for comps"))
		return
	}

	in := service.CreateOrderInput{
		LocationID:         locID,
		BaristaID:          claims.UserID,
		CustomerPhone:      req.CustomerPhone,
		LoyaltyPointsToUse: req.LoyaltyPointsToUse,
		ManualDiscountPct:  req.ManualDiscountPct,
		Comp:               req.Comp,
		CompRecipient:      req.CompRecipient,
	}

	if req.ClientUUID != "" {
		if id, err := uuid.Parse(req.ClientUUID); err == nil {
			in.ClientUUID = id
		}
	}
	if req.ShiftID != "" {
		if id, err := uuid.Parse(req.ShiftID); err == nil {
			in.ShiftID = &id
		}
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
		oi := service.OrderItemInput{
			ProductID: productID,
			Qty:       item.Qty,
		}
		if item.ClientUUID != "" {
			if id, err := uuid.Parse(item.ClientUUID); err == nil {
				oi.ClientUUID = id
			}
		}
		for _, modIDStr := range item.ModifierOptionIDs {
			modID, err := uuid.Parse(modIDStr)
			if err != nil {
				mw.Error(w, badRequestf("invalid modifier_option_id: "+modIDStr))
				return
			}
			oi.ModifierOptionIDs = append(oi.ModifierOptionIDs, modID)
		}
		in.Items = append(in.Items, oi)
	}

	for _, p := range req.Payments {
		if p.Amount <= 0 {
			mw.Error(w, badRequestf("payment amount must be > 0"))
			return
		}
		in.Payments = append(in.Payments, service.PaymentInput{
			MethodCode:  p.MethodCode,
			Amount:      p.Amount,
			ExternalRef: p.ExternalRef,
		})
	}

	order, err := h.orders.CreateOrder(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, order)
}

// POST /api/v1/orders/quote — price an order (incl. loyalty) without saving.
type quoteRequest struct {
	Items              []orderItemRequest `json:"items"`
	CustomerPhone      string             `json:"customer_phone"`
	LoyaltyPointsToUse float64            `json:"loyalty_points_to_use"`
	ManualDiscountPct  float64            `json:"manual_discount_pct"`
}

func (h *ordersHandler) quote(w http.ResponseWriter, r *http.Request) {
	var req quoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	in := service.QuoteInput{
		CustomerPhone:      req.CustomerPhone,
		LoyaltyPointsToUse: req.LoyaltyPointsToUse,
		ManualDiscountPct:  req.ManualDiscountPct,
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
		oi := service.OrderItemInput{ProductID: productID, Qty: item.Qty}
		for _, modIDStr := range item.ModifierOptionIDs {
			modID, err := uuid.Parse(modIDStr)
			if err != nil {
				mw.Error(w, badRequestf("invalid modifier_option_id: "+modIDStr))
				return
			}
			oi.ModifierOptionIDs = append(oi.ModifierOptionIDs, modID)
		}
		in.Items = append(in.Items, oi)
	}
	quote, err := h.orders.Quote(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, quote)
}

// GET /api/v1/orders/{id}
func (h *ordersHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	order, err := h.orders.Get(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, order)
}

// GET /api/v1/orders/{id}/receipt
func (h *ordersHandler) receipt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	rec, err := h.orders.Receipt(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, rec)
}

// POST /api/v1/orders/{id}/cancel
type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

func (h *ordersHandler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	claims, _ := mw.ClaimsFrom(r.Context())

	var req cancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Reason == "" {
		mw.Error(w, badRequestf("reason is required"))
		return
	}

	order, err := h.orders.Cancel(r.Context(), service.CancelOrderInput{
		OrderID:     id,
		Reason:      req.Reason,
		RequestedBy: claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, order)
}

// POST /api/v1/orders/{id}/soft-delete — hide an order from listings (admin).
func (h *ordersHandler) softDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	if err := h.orders.SoftDelete(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DELETE /api/v1/orders/{id} — permanently delete an order and reverse its
// stock/loyalty effects (admin). For cleaning up test orders.
func (h *ordersHandler) hardDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid order id"))
		return
	}
	if err := h.orders.HardDelete(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GET /api/v1/orders?shift_id=&from=&to=&limit=&offset=
func (h *ordersHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if shiftID := q.Get("shift_id"); shiftID != "" {
		id, err := uuid.Parse(shiftID)
		if err != nil {
			mw.Error(w, badRequestf("invalid shift_id"))
			return
		}
		orders, err := h.orders.ListByShift(r.Context(), id)
		if err != nil {
			mw.Error(w, err)
			return
		}
		mw.JSON(w, http.StatusOK, orders)
		return
	}

	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
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

	orders, err := h.orders.ListByLocation(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, orders)
}
