package authctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	userIDKey        ctxKey = "user_id"
	emailKey         ctxKey = "email"
	farmAuthzSkipKey ctxKey = "farm_authz_skip"
	piEdgeAuthKey    ctxKey = "pi_edge_auth"
)

func WithUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey, email)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	uid, ok := ctx.Value(userIDKey).(uuid.UUID)
	return uid, ok
}

func Email(ctx context.Context) string {
	s, _ := ctx.Value(emailKey).(string)
	return s
}

// WithFarmAuthzSkip marks the request as exempt from farm membership checks (AUTH_MODE=dev bypass only).
func WithFarmAuthzSkip(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, farmAuthzSkipKey, skip)
}

func FarmAuthzSkip(ctx context.Context) bool {
	v, _ := ctx.Value(farmAuthzSkipKey).(bool)
	return v
}

// WithPiEdgeAuth marks the request as authenticated with the shared Pi / edge API key
// (see requireJWTOrPiEdge). Farm membership checks can be skipped for scoped edge routes.
func WithPiEdgeAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, piEdgeAuthKey, true)
}

func PiEdgeAuth(ctx context.Context) bool {
	v, _ := ctx.Value(piEdgeAuthKey).(bool)
	return v
}
