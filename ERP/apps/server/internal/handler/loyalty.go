package handler

import (
	"encoding/json"
	"net/http"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type loyaltyHandler struct {
	orders *service.OrderService
}

// GET /api/v1/loyalty/config — current loyalty configuration (promo + free rule).
func (h *loyaltyHandler) config(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.orders.LoyaltyConfig(r.Context())
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, cfg)
}

type setPromoRequest struct {
	Active  bool    `json:"active"`
	Percent float64 `json:"percent"`
}

// POST /api/v1/loyalty/promo — toggle/set the global percent-off promo.
func (h *loyaltyHandler) setPromo(w http.ResponseWriter, r *http.Request) {
	var req setPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Percent < 0 || req.Percent > 100 {
		mw.Error(w, badRequestf("percent must be between 0 and 100"))
		return
	}
	cfg, err := h.orders.SetPromo(r.Context(), req.Active, req.Percent)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, cfg)
}
