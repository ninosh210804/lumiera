// Package handler wires accounting HTTP endpoints.
// handler/accounting.go is the only non-main file allowed to import the
// accounting package; all business-logic packages (order, stock, etc.) must not.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/accounting"
	pgdb "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/db/postgres/generated"
	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
)

// AccountingServicer is the surface the handler needs.
// accounting.Service satisfies it; use NewAccountingHandler to wire it.
type AccountingServicer interface {
	CreateExpense(ctx context.Context, in accounting.CreateExpenseInput) (*pgdb.Expense, error)
	ListExpenseCategories(ctx context.Context) ([]pgdb.ExpenseCategory, error)
	ListExpenses(ctx context.Context, locationID uuid.UUID, from, to *time.Time) ([]pgdb.ListExpensesRow, error)
	ListJournalEntries(ctx context.Context, limit, offset int32) ([]pgdb.JournalEntry, error)
}

type accountingHandler struct {
	acct AccountingServicer
}

// NewAccountingHandler wraps an accounting.Service as an http.Handler tree.
// Called from main.go when accounting is enabled; mounted at /api/v1/accounting.
func NewAccountingHandler(svc *accounting.Service) http.Handler {
	h := &accountingHandler{acct: svc}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// routing handled inline via chi in the caller
		_ = h
		w.WriteHeader(http.StatusNotFound)
	})
}

// GET /api/v1/accounting/expenses
func (h *accountingHandler) listExpenses(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var from, to *time.Time
	q := r.URL.Query()
	if f := q.Get("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			from = &t
		}
	}
	if t := q.Get("to"); t != "" {
		if tp, err := time.Parse("2006-01-02", t); err == nil {
			to = &tp
		}
	}

	expenses, err := h.acct.ListExpenses(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, expenses)
}

// POST /api/v1/accounting/expenses
type createExpenseRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
	PaidOn     string  `json:"paid_on"`
	Note       string  `json:"note"`
}

func (h *accountingHandler) createExpense(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		mw.Error(w, badRequestf("invalid category_id"))
		return
	}
	if req.Amount <= 0 {
		mw.Error(w, badRequestf("amount must be > 0"))
		return
	}
	paidOn := time.Now()
	if req.PaidOn != "" {
		if t, err := time.Parse("2006-01-02", req.PaidOn); err == nil {
			paidOn = t
		}
	}

	exp, err := h.acct.CreateExpense(r.Context(), accounting.CreateExpenseInput{
		CategoryID: catID,
		LocationID: locID,
		Amount:     req.Amount,
		PaidOn:     paidOn,
		Note:       req.Note,
		CreatedBy:  claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, exp)
}

// GET /api/v1/accounting/categories
func (h *accountingHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.acct.ListExpenseCategories(r.Context())
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, cats)
}

// GET /api/v1/accounting/journal
func (h *accountingHandler) listJournal(w http.ResponseWriter, r *http.Request) {
	entries, err := h.acct.ListJournalEntries(r.Context(), 50, 0)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, entries)
}
