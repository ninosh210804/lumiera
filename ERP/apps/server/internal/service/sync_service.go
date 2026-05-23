package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
)

type SyncService struct {
	q *pgdb.Queries
}

func NewSyncService(q *pgdb.Queries) *SyncService {
	return &SyncService{q: q}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type DeviceDTO struct {
	ID         uuid.UUID `json:"id"`
	LocationID uuid.UUID `json:"location_id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceName string    `json:"device_name"`
	Platform   string    `json:"platform"`
	AppVersion string    `json:"app_version"`
}

type SyncEventDTO struct {
	ID         uuid.UUID       `json:"id"`
	ClientUUID uuid.UUID       `json:"client_uuid"`
	DeviceID   uuid.UUID       `json:"device_id"`
	Sequence   int64           `json:"sequence"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	DeviceTS   time.Time       `json:"device_ts"`
	ServerTS   time.Time       `json:"server_ts"`
	Status     string          `json:"status"`
}

type PushEventInput struct {
	ClientUUID uuid.UUID
	DeviceID   uuid.UUID
	Sequence   int64
	EventType  string
	Payload    json.RawMessage
	DeviceTS   time.Time
}

type PullCursor struct {
	ServerTS time.Time
	ID       uuid.UUID
	Limit    int32
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *SyncService) RegisterDevice(ctx context.Context, locationID, userID uuid.UUID, name, platform, version string) (*DeviceDTO, error) {
	d, err := s.q.RegisterDevice(ctx, pgdb.RegisterDeviceParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		UserID:     pgtype.UUID{Bytes: userID, Valid: true},
		DeviceName: name,
		Platform:   platform,
		AppVersion: version,
	})
	if err != nil {
		return nil, err
	}
	return &DeviceDTO{
		ID:         uuid.UUID(d.ID.Bytes),
		LocationID: uuid.UUID(d.LocationID.Bytes),
		UserID:     uuid.UUID(d.UserID.Bytes),
		DeviceName: d.DeviceName,
		Platform:   d.Platform,
		AppVersion: d.AppVersion,
	}, nil
}

func (s *SyncService) Heartbeat(ctx context.Context, deviceID uuid.UUID, version string) error {
	return s.q.UpdateDeviceHeartbeat(ctx, pgdb.UpdateDeviceHeartbeatParams{
		ID:         pgtype.UUID{Bytes: deviceID, Valid: true},
		AppVersion: version,
	})
}

func (s *SyncService) PushEvent(ctx context.Context, in PushEventInput) (*SyncEventDTO, error) {
	ev, err := s.q.InsertSyncEvent(ctx, pgdb.InsertSyncEventParams{
		ClientUuid: pgtype.UUID{Bytes: in.ClientUUID, Valid: true},
		DeviceID:   pgtype.UUID{Bytes: in.DeviceID, Valid: true},
		Sequence:   in.Sequence,
		EventType:  in.EventType,
		Payload:    in.Payload,
		DeviceTs:   pgtype.Timestamptz{Time: in.DeviceTS, Valid: true},
		Status:     "pending",
	})
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows — treat as idempotent success
		return nil, nil //nolint:nilerr
	}
	return syncEventToDTO(ev), nil
}

func (s *SyncService) PullEvents(ctx context.Context, deviceID uuid.UUID, cursor PullCursor) ([]SyncEventDTO, error) {
	limit := cursor.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.q.ListSyncEventsAfterCursor(ctx, pgdb.ListSyncEventsAfterCursorParams{
		DeviceID: pgtype.UUID{Bytes: deviceID, Valid: true},
		Column2:  pgtype.Timestamptz{Time: cursor.ServerTS, Valid: true},
		Column3:  pgtype.UUID{Bytes: cursor.ID, Valid: true},
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SyncEventDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, *syncEventToDTO(r))
	}
	return out, nil
}

func (s *SyncService) GetMenuSnapshot(ctx context.Context, locationID uuid.UUID) ([]pgdb.GetMenuSnapshotRow, error) {
	return s.q.GetMenuSnapshot(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func syncEventToDTO(e pgdb.SyncEvent) *SyncEventDTO {
	dto := &SyncEventDTO{
		ID:        uuid.UUID(e.ID.Bytes),
		ClientUUID: uuid.UUID(e.ClientUuid.Bytes),
		DeviceID:  uuid.UUID(e.DeviceID.Bytes),
		Sequence:  e.Sequence,
		EventType: e.EventType,
		Payload:   e.Payload,
		Status:    e.Status,
	}
	if e.DeviceTs.Valid {
		dto.DeviceTS = e.DeviceTs.Time
	}
	if e.ServerTs.Valid {
		dto.ServerTS = e.ServerTs.Time
	}
	return dto
}
