package guardian

import (
	"fmt"
	"net/http"
	"strconv"
)

func parseFarmID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid farm id")
	}
	return id, nil
}
