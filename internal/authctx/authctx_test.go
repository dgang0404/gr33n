package authctx

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestRequestID(t *testing.T) {
	t.Parallel()
	ctx := WithRequestID(context.Background(), "abc-123")
	if RequestID(ctx) != "abc-123" {
		t.Fatalf("RequestID = %q", RequestID(ctx))
	}
	if RequestID(context.Background()) != "" {
		t.Fatal("expected empty request id")
	}
}

func TestUserIDAndEmail(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	ctx := WithEmail(WithUserID(context.Background(), uid), "dev@gr33n.local")
	got, ok := UserID(ctx)
	if !ok || got != uid {
		t.Fatalf("UserID = %v, ok=%v", got, ok)
	}
	if Email(ctx) != "dev@gr33n.local" {
		t.Fatalf("Email = %q", Email(ctx))
	}
}

func TestUserIDMissing(t *testing.T) {
	t.Parallel()
	_, ok := UserID(context.Background())
	if ok {
		t.Fatal("expected missing user id")
	}
	if Email(context.Background()) != "" {
		t.Fatal("expected empty email")
	}
}

func TestFarmAuthzSkip(t *testing.T) {
	t.Parallel()
	ctx := WithFarmAuthzSkip(context.Background(), true)
	if !FarmAuthzSkip(ctx) {
		t.Fatal("expected farm authz skip")
	}
	if FarmAuthzSkip(context.Background()) {
		t.Fatal("expected default false")
	}
}

func TestPiEdgeAuth(t *testing.T) {
	t.Parallel()
	ctx := WithPiEdgeAuth(context.Background())
	if !PiEdgeAuth(ctx) {
		t.Fatal("expected Pi edge auth")
	}
}

func TestDeviceKeyAuth(t *testing.T) {
	t.Parallel()
	ctx := WithDeviceKeyAuth(context.Background(), 42, 7)
	if !DeviceKeyAuth(ctx) {
		t.Fatal("expected device key auth")
	}
	if !PiEdgeAuth(ctx) {
		t.Fatal("device key auth should also set Pi edge auth")
	}
	rowID, ok := DeviceKeyRowID(ctx)
	if !ok || rowID != 42 {
		t.Fatalf("DeviceKeyRowID = %d, ok=%v", rowID, ok)
	}
	devID, ok := DeviceKeyDeviceID(ctx)
	if !ok || devID != 7 {
		t.Fatalf("DeviceKeyDeviceID = %d, ok=%v", devID, ok)
	}
}

func TestDeviceKeyDeviceIDInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), deviceKeyDeviceKey, int64(0))
	if DeviceKeyAuth(ctx) {
		t.Fatal("zero device id should not count as device key auth")
	}
}
