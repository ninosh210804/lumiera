package handler

import (
	"net/http"
	"strings"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type clientsHandler struct {
	orders *service.OrderService
}

// GET /api/v1/clients/profile — return a customer bonus profile by phone.
func (h *clientsHandler) profile(w http.ResponseWriter, r *http.Request) {
	phone := strings.TrimSpace(r.URL.Query().Get("phone"))
	profile, err := h.orders.CustomerProfile(r.Context(), phone)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, profile)
}
