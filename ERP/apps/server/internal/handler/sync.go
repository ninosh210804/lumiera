package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type syncHandler struct {
	sync *service.SyncService
}

// POST /api/v1/sync/register
type registerDeviceRequest struct {
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
}

func (h *syncHandler) register(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	var req registerDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.DeviceName == "" {
		mw.Error(w, badRequestf("device_name is required"))
		return
	}

	dev, err := h.sync.RegisterDevice(r.Context(), locID, claims.UserID, req.DeviceName, req.Platform, req.AppVersion)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, dev)
}

// POST /api/v1/sync/push
type pushEventsRequest struct {
	DeviceID string           `json:"device_id"`
	Events   []pushEventEntry `json:"events"`
}

type pushEventEntry struct {
	ClientUUID string          `json:"client_uuid"`
	Sequence   int64           `json:"sequence"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	DeviceTS   time.Time       `json:"device_ts"`
}

func (h *syncHandler) push(w http.ResponseWriter, r *http.Request) {
	var req pushEventsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	devID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		mw.Error(w, badRequestf("invalid device_id"))
		return
	}

	accepted := 0
	for _, e := range req.Events {
		cid, err := uuid.Parse(e.ClientUUID)
		if err != nil {
			continue
		}
		_, _ = h.sync.PushEvent(r.Context(), service.PushEventInput{
			ClientUUID: cid,
			DeviceID:   devID,
			Sequence:   e.Sequence,
			EventType:  e.EventType,
			Payload:    e.Payload,
			DeviceTS:   e.DeviceTS,
		})
		accepted++
	}
	mw.JSON(w, http.StatusOK, map[string]int{"accepted": accepted})
}

// GET /api/v1/sync/pull?device_id=&after_ts=&after_id=&limit=
func (h *syncHandler) pull(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	devID, err := uuid.Parse(q.Get("device_id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid device_id"))
		return
	}

	cursor := service.PullCursor{
		ServerTS: time.Unix(0, 0),
		Limit:    100,
	}
	if s := q.Get("after_ts"); s != "" {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			cursor.ServerTS = t
		}
	}
	if s := q.Get("after_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			cursor.ID = id
		}
	}
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			cursor.Limit = int32(n)
		}
	}

	events, err := h.sync.PullEvents(r.Context(), devID, cursor)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, events)
}

// GET /api/v1/sync/snapshot
func (h *syncHandler) snapshot(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	snap, err := h.sync.GetMenuSnapshot(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, snap)
}
