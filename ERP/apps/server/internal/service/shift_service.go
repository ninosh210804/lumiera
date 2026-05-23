package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/domain"
)

type ShiftService struct {
	q *pgdb.Queries
}

func NewShiftService(q *pgdb.Queries) *ShiftService {
	return &ShiftService{q: q}
}

// ─── DTOs ────────────────────────────────────────────────────────────────────

type ShiftDTO struct {
	ID                  uuid.UUID  `json:"id"`
	LocationID          uuid.UUID  `json:"location_id"`
	UserID              uuid.UUID  `json:"user_id"`
	UserName            string     `json:"user_name,omitempty"`
	OpenedAt            time.Time  `json:"opened_at"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	OpeningCash         float64    `json:"opening_cash"`
	ClosingCashExpected *float64   `json:"closing_cash_expected,omitempty"`
	ClosingCashActual   *float64   `json:"closing_cash_actual,omitempty"`
	Variance            *float64   `json:"variance,omitempty"`
	ClientUUID          uuid.UUID  `json:"client_uuid"`
	OrdersCount         int64      `json:"orders_count,omitempty"`
	Revenue             float64    `json:"revenue,omitempty"`
}

type CashDrawerOperationDTO struct {
	ID         uuid.UUID `json:"id"`
	ShiftID    uuid.UUID `json:"shift_id"`
	Kind       string    `json:"kind"`
	Amount     float64   `json:"amount"`
	Reason     string    `json:"reason"`
	ClientUUID uuid.UUID `json:"client_uuid"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  uuid.UUID `json:"created_by"`
}

// ─── Inputs ───────────────────────────────────────────────────────────────────

type OpenShiftInput struct {
	LocationID  uuid.UUID
	UserID      uuid.UUID
	OpeningCash float64
	ClientUUID  uuid.UUID
}

type CloseShiftInput struct {
	ShiftID             uuid.UUID
	ClosingCashExpected float64
	ClosingCashActual   float64
}

type CashDrawerInput struct {
	ShiftID    uuid.UUID
	Kind       string
	Amount     float64
	Reason     string
	ClientUUID uuid.UUID
	CreatedBy  uuid.UUID
}

// ─── Service methods ──────────────────────────────────────────────────────────

func (s *ShiftService) OpenShift(ctx context.Context, in OpenShiftInput) (*ShiftDTO, error) {
	existing, err := s.q.GetActiveShift(ctx, pgdb.GetActiveShiftParams{
		UserID:     pgtype.UUID{Bytes: in.UserID, Valid: true},
		LocationID: pgtype.UUID{Bytes: in.LocationID, Valid: true},
	})
	if err == nil && existing.ID.Valid {
		return shiftRowToDTO(existing, ""), nil
	}

	sh, err := s.q.OpenShift(ctx, pgdb.OpenShiftParams{
		LocationID:  pgtype.UUID{Bytes: in.LocationID, Valid: true},
		UserID:      pgtype.UUID{Bytes: in.UserID, Valid: true},
		OpeningCash: numericFromFloat(in.OpeningCash),
		ClientUuid:  pgtype.UUID{Bytes: in.ClientUUID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return shiftRowToDTO(sh, ""), nil
}

func (s *ShiftService) CloseShift(ctx context.Context, in CloseShiftInput) (*ShiftDTO, error) {
	sh, err := s.q.CloseShift(ctx, pgdb.CloseShiftParams{
		ID:                  pgtype.UUID{Bytes: in.ShiftID, Valid: true},
		ClosingCashExpected: numericFromFloat(in.ClosingCashExpected),
		ClosingCashActual:   numericFromFloat(in.ClosingCashActual),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("shift")
		}
		return nil, err
	}
	stats, _ := s.q.GetShiftOrderStats(ctx, sh.ID)
	dto := shiftRowToDTO(sh, "")
	applyStats(dto, stats)
	return dto, nil
}

func (s *ShiftService) GetActiveShift(ctx context.Context, userID, locationID uuid.UUID) (*ShiftDTO, error) {
	sh, err := s.q.GetActiveShift(ctx, pgdb.GetActiveShiftParams{
		UserID:     pgtype.UUID{Bytes: userID, Valid: true},
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("active shift")
		}
		return nil, err
	}
	stats, _ := s.q.GetShiftOrderStats(ctx, sh.ID)
	dto := shiftRowToDTO(sh, "")
	applyStats(dto, stats)
	return dto, nil
}

func (s *ShiftService) GetShift(ctx context.Context, shiftID uuid.UUID) (*ShiftDTO, error) {
	sh, err := s.q.GetShift(ctx, pgtype.UUID{Bytes: shiftID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("shift")
		}
		return nil, err
	}
	stats, _ := s.q.GetShiftOrderStats(ctx, sh.ID)
	dto := shiftRowToDTO(sh, "")
	applyStats(dto, stats)
	return dto, nil
}

func (s *ShiftService) ListShifts(ctx context.Context, locationID uuid.UUID) ([]ShiftDTO, error) {
	rows, err := s.q.ListShifts(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]ShiftDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, *listShiftRowToDTO(row))
	}
	return out, nil
}

func (s *ShiftService) AddCashOperation(ctx context.Context, in CashDrawerInput) (*CashDrawerOperationDTO, error) {
	op, err := s.q.CreateCashDrawerOperation(ctx, pgdb.CreateCashDrawerOperationParams{
		ShiftID:    pgtype.UUID{Bytes: in.ShiftID, Valid: true},
		Kind:       in.Kind,
		Amount:     numericFromFloat(in.Amount),
		Reason:     in.Reason,
		ClientUuid: pgtype.UUID{Bytes: in.ClientUUID, Valid: true},
		CreatedBy:  pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return cashOpToDTO(op), nil
}

func (s *ShiftService) ListCashOperations(ctx context.Context, shiftID uuid.UUID) ([]CashDrawerOperationDTO, error) {
	ops, err := s.q.ListCashDrawerOperations(ctx, pgtype.UUID{Bytes: shiftID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]CashDrawerOperationDTO, 0, len(ops))
	for _, op := range ops {
		out = append(out, *cashOpToDTO(op))
	}
	return out, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func shiftRowToDTO(sh pgdb.Shift, userName string) *ShiftDTO {
	dto := &ShiftDTO{
		ID:          uuid.UUID(sh.ID.Bytes),
		LocationID:  uuid.UUID(sh.LocationID.Bytes),
		UserID:      uuid.UUID(sh.UserID.Bytes),
		UserName:    userName,
		OpenedAt:    sh.OpenedAt.Time,
		OpeningCash: floatFromNumeric(sh.OpeningCash),
		ClientUUID:  uuid.UUID(sh.ClientUuid.Bytes),
	}
	if sh.ClosedAt.Valid {
		t := sh.ClosedAt.Time
		dto.ClosedAt = &t
	}
	if sh.ClosingCashExpected.Valid {
		v := floatFromNumeric(sh.ClosingCashExpected)
		dto.ClosingCashExpected = &v
	}
	if sh.ClosingCashActual.Valid {
		v := floatFromNumeric(sh.ClosingCashActual)
		dto.ClosingCashActual = &v
	}
	if sh.Variance.Valid {
		v := floatFromNumeric(sh.Variance)
		dto.Variance = &v
	}
	return dto
}

func listShiftRowToDTO(row pgdb.ListShiftsRow) *ShiftDTO {
	sh := pgdb.Shift{
		ID:                  row.ID,
		LocationID:          row.LocationID,
		UserID:              row.UserID,
		OpenedAt:            row.OpenedAt,
		ClosedAt:            row.ClosedAt,
		OpeningCash:         row.OpeningCash,
		ClosingCashExpected: row.ClosingCashExpected,
		ClosingCashActual:   row.ClosingCashActual,
		Variance:            row.Variance,
		AutoOpened:          row.AutoOpened,
		ClientUuid:          row.ClientUuid,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
	return shiftRowToDTO(sh, row.UserName)
}

func applyStats(dto *ShiftDTO, stats pgdb.GetShiftOrderStatsRow) {
	dto.OrdersCount = stats.OrdersCount
	if rev, ok := stats.Revenue.(pgtype.Numeric); ok {
		dto.Revenue = floatFromNumeric(rev)
	}
}

func cashOpToDTO(op pgdb.CashDrawerOperation) *CashDrawerOperationDTO {
	return &CashDrawerOperationDTO{
		ID:         uuid.UUID(op.ID.Bytes),
		ShiftID:    uuid.UUID(op.ShiftID.Bytes),
		Kind:       op.Kind,
		Amount:     floatFromNumeric(op.Amount),
		Reason:     op.Reason,
		ClientUUID: uuid.UUID(op.ClientUuid.Bytes),
		CreatedAt:  op.CreatedAt.Time,
		CreatedBy:  uuid.UUID(op.CreatedBy.Bytes),
	}
}
