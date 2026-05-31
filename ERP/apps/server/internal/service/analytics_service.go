package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
)

type AnalyticsService struct {
	q *pgdb.Queries
}

func NewAnalyticsService(q *pgdb.Queries) *AnalyticsService {
	return &AnalyticsService{q: q}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type HeatmapRow struct {
	Weekday     int32 `json:"weekday"`
	HourOfDay   int32 `json:"hour_of_day"`
	OrdersCount int64 `json:"orders_count"`
	Revenue     int64 `json:"revenue"`
	Margin      int64 `json:"margin"`
}

type ProductABCRow struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	CategoryName string    `json:"category_name"`
	BasePrice    float64   `json:"base_price"`
	UnitsSold    int64     `json:"units_sold"`
	Revenue      int64     `json:"revenue"`
	Margin       int64     `json:"margin"`
	MarginPct    float64   `json:"margin_pct"`
}

type DailyRevenueRow struct {
	Day         string  `json:"day"`
	OrdersCount int64   `json:"orders_count"`
	Revenue     int64   `json:"revenue"`
	Margin      int64   `json:"margin"`
	AvgCheck    float64 `json:"avg_check"`
	TotalCost   int64   `json:"total_cost"`
}

type BaristaStatsRow struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"full_name"`
	OrdersCount int64     `json:"orders_count"`
	Revenue     int64     `json:"revenue"`
	AvgCheck    float64   `json:"avg_check"`
	ShiftsCount int64     `json:"shifts_count"`
}

type CompSummaryRow struct {
	Recipient   string `json:"recipient"`
	OrdersCount int64  `json:"orders_count"`
	TotalValue  int64  `json:"total_value"`
	TotalCost   int64  `json:"total_cost"`
}

type CompItemRow struct {
	Recipient   string    `json:"recipient"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Qty         int64     `json:"qty"`
	TotalValue  int64     `json:"total_value"`
}

// CompReport bundles the per-recipient tally with the product breakdown of
// what each recipient took without paying (e.g. Zhandos).
type CompReport struct {
	Summary []CompSummaryRow `json:"summary"`
	Items   []CompItemRow    `json:"items"`
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *AnalyticsService) GetSalesHeatmap(ctx context.Context, locationID uuid.UUID, from, to time.Time) ([]HeatmapRow, error) {
	rows, err := s.q.GetSalesHeatmap(ctx, pgdb.GetSalesHeatmapParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		Column2:    pgtype.Timestamptz{Time: from, Valid: true},
		Column3:    pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]HeatmapRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, HeatmapRow{
			Weekday:     r.Weekday,
			HourOfDay:   r.HourOfDay,
			OrdersCount: r.OrdersCount,
			Revenue:     r.Revenue,
			Margin:      r.Margin,
		})
	}
	return out, nil
}

func (s *AnalyticsService) GetProductABC(ctx context.Context, locationID uuid.UUID) ([]ProductABCRow, error) {
	rows, err := s.q.GetProductABC(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]ProductABCRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, ProductABCRow{
			ProductID:    uuid.UUID(r.ProductID.Bytes),
			ProductName:  r.ProductName,
			CategoryName: r.CategoryName,
			BasePrice:    floatFromNumeric(r.BasePrice),
			UnitsSold:    r.UnitsSold,
			Revenue:      r.Revenue,
			Margin:       r.Margin,
			MarginPct:    floatFromNumeric(r.MarginPct),
		})
	}
	return out, nil
}

func (s *AnalyticsService) GetDailyRevenue(ctx context.Context, locationID uuid.UUID, from, to time.Time) ([]DailyRevenueRow, error) {
	rows, err := s.q.GetDailyRevenue(ctx, pgdb.GetDailyRevenueParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		Day:        pgtype.Date{Time: from, Valid: true},
		Day_2:      pgtype.Date{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]DailyRevenueRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, DailyRevenueRow{
			Day:         r.Day.Time.Format("2006-01-02"),
			OrdersCount: r.OrdersCount,
			Revenue:     r.Revenue,
			Margin:      r.Margin,
			AvgCheck:    r.AvgCheck,
			TotalCost:   r.TotalCost,
		})
	}
	return out, nil
}

func (s *AnalyticsService) GetBaristaStats(ctx context.Context, locationID uuid.UUID, from, to time.Time) ([]BaristaStatsRow, error) {
	rows, err := s.q.GetBaristaStats(ctx, pgdb.GetBaristaStatsParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		CreatedAt:  pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]BaristaStatsRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, BaristaStatsRow{
			ID:          uuid.UUID(r.ID.Bytes),
			FullName:    r.FullName,
			OrdersCount: r.OrdersCount,
			Revenue:     r.Revenue,
			AvgCheck:    r.AvgCheck,
			ShiftsCount: r.ShiftsCount,
		})
	}
	return out, nil
}

func (s *AnalyticsService) GetCompReport(ctx context.Context, locationID uuid.UUID, from, to time.Time) (*CompReport, error) {
	loc := pgtype.UUID{Bytes: locationID, Valid: true}
	fromTS := pgtype.Timestamptz{Time: from, Valid: true}
	toTS := pgtype.Timestamptz{Time: to, Valid: true}

	summaryRows, err := s.q.GetCompSummary(ctx, pgdb.GetCompSummaryParams{
		LocationID: loc, CreatedAt: fromTS, CreatedAt_2: toTS,
	})
	if err != nil {
		return nil, err
	}
	itemRows, err := s.q.GetCompItems(ctx, pgdb.GetCompItemsParams{
		LocationID: loc, CreatedAt: fromTS, CreatedAt_2: toTS,
	})
	if err != nil {
		return nil, err
	}

	rep := &CompReport{
		Summary: make([]CompSummaryRow, 0, len(summaryRows)),
		Items:   make([]CompItemRow, 0, len(itemRows)),
	}
	for _, r := range summaryRows {
		rep.Summary = append(rep.Summary, CompSummaryRow{
			Recipient:   r.CompRecipient,
			OrdersCount: r.OrdersCount,
			TotalValue:  r.TotalValue,
			TotalCost:   r.TotalCost,
		})
	}
	for _, r := range itemRows {
		rep.Items = append(rep.Items, CompItemRow{
			Recipient:   r.CompRecipient,
			ProductID:   uuid.UUID(r.ProductID.Bytes),
			ProductName: r.ProductName,
			Qty:         r.Qty,
			TotalValue:  r.TotalValue,
		})
	}
	return rep, nil
}

func (s *AnalyticsService) RefreshViews(ctx context.Context) error {
	if err := s.q.RefreshSalesHeatmap(ctx); err != nil {
		return err
	}
	if err := s.q.RefreshProductABC(ctx); err != nil {
		return err
	}
	return s.q.RefreshDailyRevenue(ctx)
}
