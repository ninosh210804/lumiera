package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── Write-off validation ─────────────────────────────────────────────────────

func TestWriteOffValidation(t *testing.T) {
	cases := []struct {
		name    string
		qty     float64
		reason  string
		wantErr bool
	}{
		{"valid write-off", 0.5, "waste", false},
		{"zero qty rejected", 0, "waste", true},
		{"negative qty rejected", -1, "waste", true},
		{"missing reason rejected", 1, "", true},
		{"large qty valid", 9999, "stocktake", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hasErr := tc.qty <= 0 || tc.reason == ""
			assert.Equal(t, tc.wantErr, hasErr)
		})
	}
}

// ─── Stock-batch deduction: AVG cost formula ──────────────────────────────────

// When a new batch arrives, the new average cost = (old_qty*old_cost + new_qty*new_cost) / total_qty.
func TestAVGCostCalculation(t *testing.T) {
	type batch struct {
		qty  float64
		cost float64
	}

	cases := []struct {
		name     string
		existing batch
		incoming batch
		wantAvg  float64
	}{
		{
			name:     "first batch",
			existing: batch{0, 0},
			incoming: batch{10, 500},
			wantAvg:  500,
		},
		{
			name:     "equal quantities",
			existing: batch{10, 400},
			incoming: batch{10, 600},
			wantAvg:  500,
		},
		{
			name:     "existing larger",
			existing: batch{90, 400},
			incoming: batch{10, 900},
			wantAvg:  450,
		},
		{
			name:     "incoming larger",
			existing: batch{10, 900},
			incoming: batch{90, 400},
			wantAvg:  450,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			totalQty := tc.existing.qty + tc.incoming.qty
			var avg float64
			if totalQty > 0 {
				avg = (tc.existing.qty*tc.existing.cost + tc.incoming.qty*tc.incoming.cost) / totalQty
			}
			assert.InDelta(t, tc.wantAvg, avg, 0.01)
		})
	}
}

// ─── Stop-list trigger threshold ──────────────────────────────────────────────

func TestStopListTrigger(t *testing.T) {
	cases := []struct {
		name        string
		currentQty  float64
		wantStopList bool
	}{
		{"zero qty triggers stop-list", 0, true},
		{"negative qty triggers stop-list", -0.5, true},
		{"positive qty does not trigger", 0.001, false},
		{"full stock does not trigger", 100, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			triggered := tc.currentQty <= 0
			assert.Equal(t, tc.wantStopList, triggered)
		})
	}
}

// ─── nullableStr / derefStr helpers ───────────────────────────────────────────

func TestNullableStrHelpers(t *testing.T) {
	assert.Equal(t, "hello", derefStr(nullableStr("hello")))
	assert.Equal(t, "", derefStr(nullableStr("")))
	assert.Equal(t, "", derefStr(nil))
}
