package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type shiftsHandler struct {
	shifts *service.ShiftService
}

// POST /api/v1/shifts/open
type openShiftRequest struct {
	OpeningCash float64   `json:"opening_cash"`
	ClientUUID  uuid.UUID `json:"client_uuid"`
}

func (h *shiftsHandler) open(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req openShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.ClientUUID == uuid.Nil {
		req.ClientUUID = uuid.New()
	}

	sh, err := h.shifts.OpenShift(r.Context(), service.OpenShiftInput{
		LocationID:  locID,
		UserID:      claims.UserID,
		OpeningCash: req.OpeningCash,
		ClientUUID:  req.ClientUUID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, sh)
}

// POST /api/v1/shifts/{id}/close
type closeShiftRequest struct {
	ClosingCashExpected float64 `json:"closing_cash_expected"`
	ClosingCashActual   float64 `json:"closing_cash_actual"`
}

func (h *shiftsHandler) close(w http.ResponseWriter, r *http.Request) {
	shiftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid shift id"))
		return
	}

	var req closeShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}

	sh, err := h.shifts.CloseShift(r.Context(), service.CloseShiftInput{
		ShiftID:             shiftID,
		ClosingCashExpected: req.ClosingCashExpected,
		ClosingCashActual:   req.ClosingCashActual,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, sh)
}

// GET /api/v1/shifts/active
func (h *shiftsHandler) active(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	sh, err := h.shifts.GetActiveShift(r.Context(), claims.UserID, locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, sh)
}

// GET /api/v1/shifts/{id}
func (h *shiftsHandler) get(w http.ResponseWriter, r *http.Request) {
	shiftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid shift id"))
		return
	}
	sh, err := h.shifts.GetShift(r.Context(), shiftID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, sh)
}

// GET /api/v1/shifts
func (h *shiftsHandler) list(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	shifts, err := h.shifts.ListShifts(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, shifts)
}

// POST /api/v1/shifts/{id}/cash
type cashDrawerRequest struct {
	Kind       string    `json:"kind"`
	Amount     float64   `json:"amount"`
	Reason     string    `json:"reason"`
	ClientUUID uuid.UUID `json:"client_uuid"`
}

func (h *shiftsHandler) addCash(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	shiftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid shift id"))
		return
	}

	var req cashDrawerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Kind == "" {
		mw.Error(w, badRequestf("kind is required"))
		return
	}
	if req.Amount <= 0 {
		mw.Error(w, badRequestf("amount must be > 0"))
		return
	}
	if req.ClientUUID == uuid.Nil {
		req.ClientUUID = uuid.New()
	}

	op, err := h.shifts.AddCashOperation(r.Context(), service.CashDrawerInput{
		ShiftID:    shiftID,
		Kind:       req.Kind,
		Amount:     req.Amount,
		Reason:     req.Reason,
		ClientUUID: req.ClientUUID,
		CreatedBy:  claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, op)
}

// GET /api/v1/shifts/{id}/cash
func (h *shiftsHandler) listCash(w http.ResponseWriter, r *http.Request) {
	shiftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid shift id"))
		return
	}
	ops, err := h.shifts.ListCashOperations(r.Context(), shiftID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, ops)
}
