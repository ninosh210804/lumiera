package service

// Unit conversion for recipe / inventory math.
//
// Ingredients store qty + avg cost in their own unit (g/kg/ml/l/pcs). Recipe
// items can specify a different but compatible unit (e.g. ingredient in "l",
// recipe needs "ml"). Without conversion, qty * unit_cost is off by a factor
// of 1000 — which is the root of the "себестоимость не всегда считается
// правильно" report (/3).
//
// All weights normalise to grams, volumes to millilitres, counts stay 1:1.
// Incompatible units (e.g. recipe in "g" against an ingredient stored in
// "ml") return the input qty unchanged so existing data isn't silently
// mangled — caller treats it as best-effort.

var unitToBase = map[string]float64{
	"g":   1,
	"kg":  1000,
	"ml":  1,
	"l":   1000,
	"pcs": 1,
}

var unitCategory = map[string]string{
	"g":   "weight",
	"kg":  "weight",
	"ml":  "volume",
	"l":   "volume",
	"pcs": "count",
}

// convertQty converts qty from one unit to another. If either unit is unknown
// or the categories differ, returns qty unchanged.
func convertQty(qty float64, fromUnit, toUnit string) float64 {
	if fromUnit == toUnit || fromUnit == "" || toUnit == "" {
		return qty
	}
	fromFactor, ok1 := unitToBase[fromUnit]
	toFactor, ok2 := unitToBase[toUnit]
	if !ok1 || !ok2 || unitCategory[fromUnit] != unitCategory[toUnit] {
		return qty
	}
	return qty * fromFactor / toFactor
}
