package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/domain"
	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type usersHandler struct {
	users *service.UserService
}

// --- GET /api/v1/roles ---

func (h *usersHandler) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.users.ListRoles(r.Context())
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, roles)
}

// --- GET /api/v1/users?location_id=... ---

func (h *usersHandler) list(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	var locationID uuid.UUID
	if locStr != "" {
		var err error
		locationID, err = uuid.Parse(locStr)
		if err != nil {
			mw.Error(w, badRequestf("invalid location_id"))
			return
		}
	} else {
		// Fall back to the location from the authenticated user's token
		claims, ok := mw.ClaimsFrom(r.Context())
		if !ok {
			mw.Error(w, domain.NewUnauthorized("authentication required"))
			return
		}
		locationID = claims.LocationID
	}

	items, err := h.users.ListByLocation(r.Context(), locationID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, items)
}

// --- POST /api/v1/users ---

type createUserRequest struct {
	LocationID string  `json:"location_id"`
	RoleID     string  `json:"role_id"`
	FullName   string  `json:"full_name"`
	Email      *string `json:"email"`
	PIN        string  `json:"pin"`
}

func (h *usersHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		mw.Error(w, badRequestf("invalid location_id"))
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		mw.Error(w, badRequestf("invalid role_id"))
		return
	}
	if req.FullName == "" {
		mw.Error(w, badRequestf("full_name is required"))
		return
	}
	if len(req.PIN) < 4 || len(req.PIN) > 6 {
		mw.Error(w, badRequestf("pin must be 4–6 digits"))
		return
	}

	claims, _ := mw.ClaimsFrom(r.Context())
	user, err := h.users.Create(r.Context(), service.CreateUserRequest{
		LocationID: locationID,
		RoleID:     roleID,
		FullName:   req.FullName,
		Email:      req.Email,
		PIN:        req.PIN,
		CreatedBy:  claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, user)
}

// --- PUT /api/v1/users/{id}/pin ---

type updatePINRequest struct {
	PIN string `json:"pin"`
}

func (h *usersHandler) updatePIN(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid user id"))
		return
	}

	// Only admin or the user themselves may change the PIN
	claims, _ := mw.ClaimsFrom(r.Context())
	if claims.Role != domain.RoleAdmin && claims.UserID != userID {
		mw.Error(w, domain.ErrInsufficientPerm)
		return
	}

	var req updatePINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if len(req.PIN) < 4 || len(req.PIN) > 6 {
		mw.Error(w, badRequestf("pin must be 4–6 digits"))
		return
	}

	if err := h.users.UpdatePIN(r.Context(), userID, req.PIN); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- DELETE /api/v1/users/{id} ---

func (h *usersHandler) deactivate(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid user id"))
		return
	}
	if err := h.users.Deactivate(r.Context(), userID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
