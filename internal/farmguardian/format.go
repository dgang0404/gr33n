// Phase 28 WS3 — number formatters used by the cycle-analytics snapshot
// renderer. Kept in a separate file so unit tests can target them
// directly without touching the SQL path.

package farmguardian

import (
	"math"
	"strconv"
)

// formatInt prints an int64 with no decimals.
func formatInt(v int64) string { return strconv.FormatInt(v, 10) }

// formatLiters prints liters with up to 1 decimal — Guardian doesn't need
// sub-100mL precision in a prompt summary. Whole-liter values render
// without a decimal so "980L" not "980.0L".
func formatLiters(v float64) string {
	r := math.Round(v*10) / 10
	if r == math.Trunc(r) {
		return strconv.FormatInt(int64(r), 10) + "L"
	}
	return strconv.FormatFloat(r, 'f', 1, 64) + "L"
}

// formatLitersPerDay shows ".../d" with 1 decimal.
func formatLitersPerDay(v float64) string {
	return strconv.FormatFloat(math.Round(v*10)/10, 'f', 1, 64) + "L/d"
}

// formatEC shows EC with 2 decimals (the canonical resolution operators
// use when setting recipe targets).
func formatEC(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
}

// formatPH shows pH with 2 decimals.
func formatPH(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
}

// formatMoney emits 2 decimals — currency code is appended by the caller.
func formatMoney(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
}

// formatGrams shows grams with no decimals (a gram is fine resolution
// for a cycle-level yield summary).
func formatGrams(v float64) string {
	return strconv.FormatInt(int64(math.Round(v)), 10)
}

// formatGramsPerDay shows ".../d" with 2 decimals — yield-rate is the
// most often-cited efficiency metric and the operator wants the trailing
// fraction.
func formatGramsPerDay(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64) + "g/d"
}
