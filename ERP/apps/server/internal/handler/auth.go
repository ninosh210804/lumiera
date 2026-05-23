package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type authHandler struct {
	auth *service.AuthService
}

// --- POST /api/v1/auth/login ---
// Admin / manager: { "email": "...", "pin": "1234" }

type loginEmailRequest struct {
	Email string `json:"email"`
	PIN   string `json:"pin"`
}

func (h *authHandler) loginEmail(w http.ResponseWriter, r *http.Request) {
	var req loginEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Email == "" || req.PIN == "" {
		mw.Error(w, badRequestf("email and pin are required"))
		return
	}

	resp, err := h.auth.LoginEmail(r.Context(), req.Email, req.PIN)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, resp)
}

// --- POST /api/v1/auth/pin ---
// Barista: { "user_id": "...", "location_id": "...", "pin": "1234" }

type loginPINRequest struct {
	UserID     string `json:"user_id"`
	LocationID string `json:"location_id"`
	PIN        string `json:"pin"`
}

func (h *authHandler) loginPIN(w http.ResponseWriter, r *http.Request) {
	var req loginPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		mw.Error(w, badRequestf("invalid user_id"))
		return
	}
	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		mw.Error(w, badRequestf("invalid location_id"))
		return
	}
	if req.PIN == "" {
		mw.Error(w, badRequestf("pin is required"))
		return
	}

	resp, err := h.auth.LoginPIN(r.Context(), userID, locationID, req.PIN)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, resp)
}

// --- GET /api/v1/auth/baristas?location_id=... ---
// Returns the barista list for the PIN-selection screen (no auth required).

func (h *authHandler) listBaristas(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	if locStr == "" {
		mw.Error(w, badRequestf("location_id query param is required"))
		return
	}
	locationID, err := uuid.Parse(locStr)
	if err != nil {
		mw.Error(w, badRequestf("invalid location_id"))
		return
	}

	items, err := h.auth.ListBaristas(r.Context(), locationID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, items)
}

// --- GET /api/v1/auth/me --- (requires JWT)

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	info, err := h.auth.GetMe(r.Context(), claims.UserID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, info)
}
