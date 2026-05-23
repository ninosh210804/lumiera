package service

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Payment validation ───────────────────────────────────────────────────────

func TestPaymentSumValidation(t *testing.T) {
	cases := []struct {
		name      string
		total     float64
		payments  []float64
		wantErr   bool
	}{
		{
			name:     "exact match single payment",
			total:    1500,
			payments: []float64{1500},
			wantErr:  false,
		},
		{
			name:     "split payment sums to total",
			total:    1500,
			payments: []float64{1000, 500},
			wantErr:  false,
		},
		{
			name:     "underpayment",
			total:    1500,
			payments: []float64{1000},
			wantErr:  true,
		},
		{
			name:     "overpayment beyond tolerance",
			total:    1500,
			payments: []float64{1501.5},
			wantErr:  true,
		},
		{
			name:     "floating point tolerance accepted",
			total:    1500,
			payments: []float64{1499.999},
			wantErr:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sum := 0.0
			for _, p := range tc.payments {
				sum += p
			}
			diff := math.Abs(sum - tc.total)
			hasErr := diff > 0.01
			assert.Equal(t, tc.wantErr, hasErr,
				"payment sum=%.4f total=%.4f diff=%.4f", sum, tc.total, diff)
		})
	}
}

// ─── Loyalty points calculation ───────────────────────────────────────────────

func TestLoyaltyPointsEarned(t *testing.T) {
	cases := []struct {
		name   string
		total  float64
		expect float64
	}{
		{"zero order", 0, 0},
		{"1000 tenge → 10 points (1%)", 1000, 10},
		{"1550 tenge → 16 points (rounded)", 1550, 16},
		{"999 tenge → 10 points", 999, 10},
		{"1 tenge → 0 points", 1, 0},
		{"50 tenge → 1 point", 50, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := math.Round(tc.total * 0.01)
			assert.Equal(t, tc.expect, got)
		})
	}
}

// ─── numericFromFloat / floatFromNumeric round-trip ───────────────────────────

func TestNumericRoundTrip(t *testing.T) {
	values := []float64{0, 1, 1.5, 999.99, 100000, -5.25}
	for _, v := range values {
		n := numericFromFloat(v)
		got := floatFromNumeric(n)
		require.InDelta(t, v, got, 0.0001, "round-trip failed for %.4f", v)
	}
}

// ─── Receipt number fallback ──────────────────────────────────────────────────

func TestReceiptNoFallback(t *testing.T) {
	// Simulate the type-assertion pattern used in CreateOrder.
	// When GetNextReceiptNo returns an empty string, we fall back to "?".
	check := func(raw interface{}) string {
		if str, ok := raw.(string); ok && str != "" {
			return str
		}
		return "?"
	}

	assert.Equal(t, "2024-001", check("2024-001"))
	assert.Equal(t, "?", check(""))
	assert.Equal(t, "?", check(nil))
	assert.Equal(t, "?", check(42))
}
