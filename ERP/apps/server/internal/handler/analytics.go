package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	mw "github.com/ninosh210804/lumiera/apps/server/internal/middleware"
	"github.com/ninosh210804/lumiera/apps/server/internal/service"
)

type analyticsHandler struct {
	analytics *service.AnalyticsService
}

// GET /api/v1/analytics/heatmap?from=2006-01-02&to=2006-01-02
func (h *analyticsHandler) heatmap(w http.ResponseWriter, r *http.Request) {
	locID, from, to := analyticsParams(r)
	data, err := h.analytics.GetSalesHeatmap(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, data)
}

// GET /api/v1/analytics/products
func (h *analyticsHandler) productABC(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}
	data, err := h.analytics.GetProductABC(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, data)
}

// GET /api/v1/analytics/revenue?from=2006-01-02&to=2006-01-02
func (h *analyticsHandler) dailyRevenue(w http.ResponseWriter, r *http.Request) {
	locID, from, to := analyticsParams(r)
	data, err := h.analytics.GetDailyRevenue(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, data)
}

// GET /api/v1/analytics/baristas?from=2006-01-02&to=2006-01-02
func (h *analyticsHandler) baristaStats(w http.ResponseWriter, r *http.Request) {
	locID, from, to := analyticsParams(r)
	data, err := h.analytics.GetBaristaStats(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, data)
}

// GET /api/v1/analytics/comps?from=2006-01-02&to=2006-01-02
// Products taken without payment (e.g. Zhandos), per recipient + breakdown.
func (h *analyticsHandler) comps(w http.ResponseWriter, r *http.Request) {
	locID, from, to := analyticsParams(r)
	data, err := h.analytics.GetCompReport(r.Context(), locID, from, to)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, data)
}

// POST /api/v1/analytics/refresh  (admin only)
func (h *analyticsHandler) refreshViews(w http.ResponseWriter, r *http.Request) {
	if err := h.analytics.RefreshViews(r.Context()); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "refreshed"})
}

// analyticsParams extracts location_id + date range from request.
func analyticsParams(r *http.Request) (locID uuid.UUID, from, to time.Time) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		locID = claims.LocationID
	}

	q := r.URL.Query()
	from = time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, -1, 0)
	to = time.Now().UTC()
	if f := q.Get("from"); f != "" {
		if t, ok := parseDateOrTime(f); ok {
			from = t
		}
	}
	if t := q.Get("to"); t != "" {
		if tp, ok := parseDateOrTime(t); ok {
			// If only a date was provided (no time component), extend to end of day.
			if tp.Hour() == 0 && tp.Minute() == 0 && tp.Second() == 0 && len(t) <= 10 {
				to = tp.Add(24 * time.Hour)
			} else {
				to = tp
			}
		}
	}
	return locID, from, to
}

// parseDateOrTime accepts either "2006-01-02" date or RFC3339 timestamps.
func parseDateOrTime(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, true
	}
	return time.Time{}, false
}
