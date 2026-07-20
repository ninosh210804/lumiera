package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
)

// Rule codes read by the pricing engine.
const (
	rulePromoDiscount = "promo_discount"
	ruleFreeEveryN    = "free_every_n"
	rulePointsEarn    = "earn_1pct"
)

// LoyaltyConfig is the effective, resolved loyalty configuration for an order.
type LoyaltyConfig struct {
	PromoActive  bool    `json:"promo_active"`
	PromoPct     float64 `json:"promo_percent"`
	EveryN       int     `json:"every_n"`
	FreeCategory string  `json:"free_category"`
	EarnPct      float64 `json:"earn_percent"`
}

// loadLoyaltyConfig resolves rule rows into a LoyaltyConfig with sane defaults.
func loadLoyaltyConfig(ctx context.Context, q *pgdb.Queries) (LoyaltyConfig, error) {
	cfg := LoyaltyConfig{EveryN: 7, FreeCategory: "Кофе", PromoPct: 10, EarnPct: 5}
	rules, err := q.GetLoyaltyRules(ctx)
	if err != nil {
		return cfg, err
	}
	for _, r := range rules {
		var p struct {
			Percent  *float64 `json:"percent"`
			EveryN   *int     `json:"every_n"`
			Category *string  `json:"category"`
		}
		_ = json.Unmarshal(r.Params, &p)
		switch r.Code {
		case rulePromoDiscount:
			cfg.PromoActive = r.IsActive
			if p.Percent != nil {
				cfg.PromoPct = *p.Percent
			}
		case ruleFreeEveryN:
			if !r.IsActive {
				cfg.EveryN = 0 // disabled
			} else if p.EveryN != nil {
				cfg.EveryN = *p.EveryN
			}
			if p.Category != nil && *p.Category != "" {
				cfg.FreeCategory = *p.Category
			}
		case rulePointsEarn:
			if r.IsActive && p.Percent != nil {
				cfg.EarnPct = *p.Percent
			} else if !r.IsActive {
				cfg.EarnPct = 0
			}
		}
	}
	return cfg, nil
}

// loyaltyState is a customer's loyalty account snapshot used for pricing.
type loyaltyState struct {
	found          bool
	customerID     [16]byte
	pointsBalance  float64
	freeDrinksLeft int
	coffeePunches  int
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}

	var digits strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	value := digits.String()
	if len(value) == 11 && strings.HasPrefix(value, "8") {
		value = "7" + value[1:]
	}
	return value
}

// lookupLoyalty fetches a customer's account by phone without creating anything
// (used by quotes). Returns found=false when the customer is new.
func lookupLoyalty(ctx context.Context, q *pgdb.Queries, phone string) (loyaltyState, error) {
	var st loyaltyState
	phone = normalizePhone(phone)
	cust, err := q.GetCustomerByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return st, nil
		}
		return st, err
	}
	acct, err := q.GetLoyaltyAccountByCustomer(ctx, cust.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.found = true
			st.customerID = cust.ID.Bytes
			return st, nil
		}
		return st, err
	}
	st.found = true
	st.customerID = cust.ID.Bytes
	st.pointsBalance = floatFromNumeric(acct.PointsBalance)
	st.freeDrinksLeft = int(acct.FreeDrinksLeft)
	st.coffeePunches = int(acct.CoffeePunches)
	return st, nil
}

// discountResult is the outcome of applying loyalty rules to an order.
type discountResult struct {
	PromoDiscount      float64
	ManualDiscount     float64 // per-order barista discount (e.g. paper ticket)
	FreeCoffeeDiscount float64
	FreeCoffeesApplied int
	PointsUsed         float64
	DiscountTotal      float64 // promo + manual + free coffee
	Total              float64
	// post-order account state (only meaningful for registered customers)
	NewFreeDrinksLeft int
	NewCoffeePunches  int
	EarnedPoints      float64
}

// computeDiscounts applies the free-Nth-coffee and global promo rules.
//
//	subtotal:         sum of line totals
//	coffeeUnitPrices: per-unit prices of qualifying ("coffee") units in the cart
//	st:               customer loyalty snapshot (zero value = walk-in)
//	pointsToUse:      points the customer wants to redeem (1 point = 1 ₸)
//	manualPct:        per-order barista discount percent (e.g. paper ticket); 0 = none
func computeDiscounts(subtotal float64, coffeeUnitPrices []float64, st loyaltyState, cfg LoyaltyConfig, pointsToUse, manualPct float64) discountResult {
	var res discountResult

	// Free coffee: redeem drinks earned on previous visits, applied to the most
	// expensive qualifying units first.
	prices := append([]float64(nil), coffeeUnitPrices...)
	sort.Sort(sort.Reverse(sort.Float64Slice(prices)))
	freeApplied := 0
	if st.freeDrinksLeft > 0 && len(prices) > 0 {
		freeApplied = st.freeDrinksLeft
		if freeApplied > len(prices) {
			freeApplied = len(prices)
		}
		for i := 0; i < freeApplied; i++ {
			res.FreeCoffeeDiscount += prices[i]
		}
	}
	res.FreeCoffeesApplied = freeApplied
	res.FreeCoffeeDiscount = round2(res.FreeCoffeeDiscount)

	// Punch counter advances on paid qualifying units; earns new free drinks.
	paidCoffees := len(prices) - freeApplied
	earnedFree := 0
	newPunches := st.coffeePunches
	if cfg.EveryN > 0 {
		newPunches += paidCoffees
		earnedFree = newPunches / cfg.EveryN
		newPunches = newPunches % cfg.EveryN
	}
	res.NewCoffeePunches = newPunches
	res.NewFreeDrinksLeft = st.freeDrinksLeft - freeApplied + earnedFree

	// Promo + manual (ticket) discount: percent off the payable amount (after
	// free coffee). Both apply to the same base so a customer with a ticket still
	// benefits from an active global promo.
	payable := subtotal - res.FreeCoffeeDiscount
	if payable > 0 {
		if cfg.PromoActive && cfg.PromoPct > 0 {
			res.PromoDiscount = round2(payable * cfg.PromoPct / 100)
		}
		if manualPct > 0 {
			res.ManualDiscount = round2(payable * manualPct / 100)
		}
	}

	res.DiscountTotal = round2(res.PromoDiscount + res.ManualDiscount + res.FreeCoffeeDiscount)
	if res.DiscountTotal > subtotal {
		res.DiscountTotal = subtotal // never discount below zero
	}

	// Points redemption, capped by balance and remaining payable amount.
	remaining := subtotal - res.DiscountTotal
	if pointsToUse > 0 && st.pointsBalance > 0 && remaining > 0 {
		used := math.Min(pointsToUse, st.pointsBalance)
		used = math.Min(used, remaining)
		res.PointsUsed = round2(used)
	}

	res.Total = round2(subtotal - res.DiscountTotal - res.PointsUsed)
	if res.Total < 0 {
		res.Total = 0
	}

	// Points earned on the amount actually paid.
	if cfg.EarnPct > 0 {
		res.EarnedPoints = math.Round(res.Total * cfg.EarnPct / 100)
	}
	return res
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
