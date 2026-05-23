// Package accounting implements double-entry bookkeeping for the ERP.
// It is imported ONLY from cmd/server/main.go and activated via cfg.Accounting.Enabled.
// All other packages must NOT import this package.
package accounting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/eventbus"
)

// Service listens to the event bus and creates journal entries.
type Service struct {
	q   *pgdb.Queries
	bus *eventbus.Bus
}

// New creates the accounting service and subscribes to relevant events.
func New(q *pgdb.Queries, bus *eventbus.Bus) *Service {
	svc := &Service{q: q, bus: bus}
	bus.Subscribe(eventbus.EventSalePaid, svc.onSalePaid)
	bus.Subscribe(eventbus.EventStockReceived, svc.onStockReceived)
	return svc
}

// ─── Event handlers ───────────────────────────────────────────────────────────

func (s *Service) onSalePaid(e eventbus.Event) {
	p, ok := e.Payload.(eventbus.SalePaidPayload)
	if !ok {
		return
	}
	ctx := context.Background()

	cashAcc, err := s.q.GetAccountByCode(ctx, "1010")   // Касса
	revenueAcc, err2 := s.q.GetAccountByCode(ctx, "5010") // Выручка
	cogsAcc, err3 := s.q.GetAccountByCode(ctx, "7010")    // Себестоимость
	stockAcc, err4 := s.q.GetAccountByCode(ctx, "1310")   // Товары на складе
	if err != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Printf("accounting: account lookup failed for sale %s", p.OrderID)
		return
	}

	orderID, _ := uuid.Parse(p.OrderID)
	entry, err := s.q.CreateJournalEntry(ctx, pgdb.CreateJournalEntryParams{
		PostedOn:  pgtype.Date{Time: e.OccurredAt, Valid: true},
		Memo:      fmt.Sprintf("Sale %s", p.OrderID),
		Source:    "sale",
		SourceRef: pgtype.UUID{Bytes: orderID, Valid: true},
	})
	if err != nil {
		log.Printf("accounting: create journal entry failed: %v", err)
		return
	}

	jid := entry.ID
	lines := []pgdb.AddJournalLineParams{
		{JournalEntryID: jid, AccountID: cashAcc.ID, Debit: numericFromFloat(p.Total), Credit: zero(), Memo: "cash in"},
		{JournalEntryID: jid, AccountID: revenueAcc.ID, Debit: zero(), Credit: numericFromFloat(p.Total), Memo: "revenue"},
		{JournalEntryID: jid, AccountID: cogsAcc.ID, Debit: numericFromFloat(p.CostTotal), Credit: zero(), Memo: "COGS"},
		{JournalEntryID: jid, AccountID: stockAcc.ID, Debit: zero(), Credit: numericFromFloat(p.CostTotal), Memo: "stock out"},
	}
	for _, l := range lines {
		if _, err := s.q.AddJournalLine(ctx, l); err != nil {
			log.Printf("accounting: add journal line failed: %v", err)
		}
	}
}

func (s *Service) onStockReceived(e eventbus.Event) {
	p, ok := e.Payload.(eventbus.StockReceivedPayload)
	if !ok {
		return
	}
	ctx := context.Background()

	stockAcc, err := s.q.GetAccountByCode(ctx, "1310")
	apAcc, err2 := s.q.GetAccountByCode(ctx, "3010") // Кредиторская задолженность
	if err != nil || err2 != nil {
		log.Printf("accounting: account lookup failed for receipt %s", p.PurchaseOrderID)
		return
	}

	poID, _ := uuid.Parse(p.PurchaseOrderID)
	entry, err := s.q.CreateJournalEntry(ctx, pgdb.CreateJournalEntryParams{
		PostedOn:  pgtype.Date{Time: e.OccurredAt, Valid: true},
		Memo:      fmt.Sprintf("Stock receipt PO %s", p.PurchaseOrderID),
		Source:    "purchase",
		SourceRef: pgtype.UUID{Bytes: poID, Valid: true},
	})
	if err != nil {
		log.Printf("accounting: create journal entry failed: %v", err)
		return
	}

	jid := entry.ID
	lines := []pgdb.AddJournalLineParams{
		{JournalEntryID: jid, AccountID: stockAcc.ID, Debit: numericFromFloat(p.TotalAmount), Credit: zero(), Memo: "stock in"},
		{JournalEntryID: jid, AccountID: apAcc.ID, Debit: zero(), Credit: numericFromFloat(p.TotalAmount), Memo: "AP"},
	}
	for _, l := range lines {
		if _, err := s.q.AddJournalLine(ctx, l); err != nil {
			log.Printf("accounting: add journal line failed: %v", err)
		}
	}
}

// ─── Expense operations (called from HTTP handler) ────────────────────────────

type CreateExpenseInput struct {
	CategoryID uuid.UUID
	LocationID uuid.UUID
	Amount     float64
	PaidOn     time.Time
	Note       string
	CreatedBy  uuid.UUID
}

func (s *Service) CreateExpense(ctx context.Context, in CreateExpenseInput) (*pgdb.Expense, error) {
	exp, err := s.q.CreateExpense(ctx, pgdb.CreateExpenseParams{
		ExpenseCategoryID: pgtype.UUID{Bytes: in.CategoryID, Valid: true},
		LocationID:        pgtype.UUID{Bytes: in.LocationID, Valid: true},
		Amount:            numericFromFloat(in.Amount),
		PaidOn:            pgtype.Date{Time: in.PaidOn, Valid: true},
		Note:              in.Note,
		CreatedBy:         pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Auto-post journal entry for the expense
	expAcc, err := s.q.GetAccountByCode(ctx, "7210") // Административные расходы
	cashAcc, err2 := s.q.GetAccountByCode(ctx, "1010")
	if err == nil && err2 == nil {
		entry, err := s.q.CreateJournalEntry(ctx, pgdb.CreateJournalEntryParams{
			PostedOn:  pgtype.Date{Time: in.PaidOn, Valid: true},
			Memo:      in.Note,
			Source:    "manual",
			SourceRef: exp.ID,
		})
		if err == nil {
			jid := entry.ID
			_ = s.addLine(ctx, jid, expAcc.ID, in.Amount, 0, "expense")
			_ = s.addLine(ctx, jid, cashAcc.ID, 0, in.Amount, "cash out")
			_ = s.q.UpdateExpenseJournalRef(ctx, pgdb.UpdateExpenseJournalRefParams{
				ID:             exp.ID,
				JournalEntryID: entry.ID,
			})
		}
	}
	return &exp, nil
}

func (s *Service) ListExpenseCategories(ctx context.Context) ([]pgdb.ExpenseCategory, error) {
	return s.q.ListExpenseCategories(ctx)
}

func (s *Service) ListExpenses(ctx context.Context, locationID uuid.UUID, from, to *time.Time) ([]pgdb.ListExpensesRow, error) {
	var fromDate, toDate pgtype.Date
	if from != nil {
		fromDate = pgtype.Date{Time: *from, Valid: true}
	}
	if to != nil {
		toDate = pgtype.Date{Time: *to, Valid: true}
	}
	return s.q.ListExpenses(ctx, pgdb.ListExpensesParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		Column2:    fromDate,
		Column3:    toDate,
	})
}

func (s *Service) ListJournalEntries(ctx context.Context, limit, offset int32) ([]pgdb.JournalEntry, error) {
	return s.q.ListJournalEntries(ctx, pgdb.ListJournalEntriesParams{
		Limit:  limit,
		Offset: offset,
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *Service) addLine(ctx context.Context, jeID, accID pgtype.UUID, debit, credit float64, memo string) error {
	_, err := s.q.AddJournalLine(ctx, pgdb.AddJournalLineParams{
		JournalEntryID: jeID,
		AccountID:      accID,
		Debit:          numericFromFloat(debit),
		Credit:         numericFromFloat(credit),
		Memo:           memo,
	})
	return err
}

func numericFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.4f", f))
	return n
}

func zero() pgtype.Numeric {
	return numericFromFloat(0)
}
