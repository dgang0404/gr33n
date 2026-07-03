package catalognotify

import (
	"context"
	"testing"
)

func TestSync_NilQuerier(t *testing.T) {
	_, err := Sync(context.Background(), nil, nil)
	if err == nil || err.Error() != "catalognotify: nil querier" {
		t.Fatalf("err = %v", err)
	}
}
