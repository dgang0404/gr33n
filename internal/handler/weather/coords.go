package weather

import (
	"fmt"
	"strconv"
)

func ifaceFloat(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		s := fmt.Sprint(x)
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	}
}
