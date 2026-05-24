package service

import "testing"

func TestComputeDiscounts_PromoOnly(t *testing.T) {
	cfg := LoyaltyConfig{PromoActive: true, PromoPct: 10, EveryN: 7, FreeCategory: "Кофе", EarnPct: 1}
	res := computeDiscounts(2000, []float64{1000, 1000}, loyaltyState{}, cfg, 0)
	if res.PromoDiscount != 200 {
		t.Fatalf("promo: want 200, got %v", res.PromoDiscount)
	}
	if res.FreeCoffeeDiscount != 0 {
		t.Fatalf("free: want 0, got %v", res.FreeCoffeeDiscount)
	}
	if res.Total != 1800 {
		t.Fatalf("total: want 1800, got %v", res.Total)
	}
}

func TestComputeDiscounts_FreeCoffeeUsesMostExpensive(t *testing.T) {
	cfg := LoyaltyConfig{EveryN: 7, FreeCategory: "Кофе"}
	st := loyaltyState{found: true, freeDrinksLeft: 1}
	res := computeDiscounts(2200, []float64{1200, 1000}, st, cfg, 0)
	if res.FreeCoffeesApplied != 1 {
		t.Fatalf("applied: want 1, got %d", res.FreeCoffeesApplied)
	}
	if res.FreeCoffeeDiscount != 1200 { // most expensive
		t.Fatalf("free discount: want 1200, got %v", res.FreeCoffeeDiscount)
	}
	if res.Total != 1000 {
		t.Fatalf("total: want 1000, got %v", res.Total)
	}
	if res.NewFreeDrinksLeft != 0 {
		t.Fatalf("new free left: want 0, got %d", res.NewFreeDrinksLeft)
	}
}

func TestComputeDiscounts_PunchEarnsFreeAfterN(t *testing.T) {
	cfg := LoyaltyConfig{EveryN: 7, FreeCategory: "Кофе"}
	// 6 punches + 1 paid coffee = 7 -> earns 1 free, punches reset to 0.
	st := loyaltyState{found: true, freeDrinksLeft: 0, coffeePunches: 6}
	res := computeDiscounts(800, []float64{800}, st, cfg, 0)
	if res.FreeCoffeesApplied != 0 {
		t.Fatalf("no free drink available yet, got %d", res.FreeCoffeesApplied)
	}
	if res.NewFreeDrinksLeft != 1 {
		t.Fatalf("earned free: want 1, got %d", res.NewFreeDrinksLeft)
	}
	if res.NewCoffeePunches != 0 {
		t.Fatalf("punches reset: want 0, got %d", res.NewCoffeePunches)
	}
}

func TestComputeDiscounts_PromoAfterFreeCoffee(t *testing.T) {
	cfg := LoyaltyConfig{PromoActive: true, PromoPct: 10, EveryN: 7, FreeCategory: "Кофе"}
	st := loyaltyState{found: true, freeDrinksLeft: 1}
	// subtotal 2000 (two 1000 coffees); one free -> payable 1000; promo 10% -> 100.
	res := computeDiscounts(2000, []float64{1000, 1000}, st, cfg, 0)
	if res.FreeCoffeeDiscount != 1000 {
		t.Fatalf("free: want 1000, got %v", res.FreeCoffeeDiscount)
	}
	if res.PromoDiscount != 100 {
		t.Fatalf("promo on payable: want 100, got %v", res.PromoDiscount)
	}
	if res.Total != 900 {
		t.Fatalf("total: want 900, got %v", res.Total)
	}
}

func TestComputeDiscounts_WalkInNoPromo(t *testing.T) {
	cfg := LoyaltyConfig{PromoActive: false, EveryN: 7, FreeCategory: "Кофе", EarnPct: 1}
	res := computeDiscounts(1500, []float64{1500}, loyaltyState{}, cfg, 0)
	if res.DiscountTotal != 0 || res.Total != 1500 {
		t.Fatalf("walk-in: want total 1500 no discount, got total=%v disc=%v", res.Total, res.DiscountTotal)
	}
}
