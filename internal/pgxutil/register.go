package pgxutil

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

// RegisterVectorTypes configures AfterConnect so pgvector types decode for pgx/sqlc.
func RegisterVectorTypes(cfg *pgxpool.Config) {
	if cfg == nil {
		return
	}
	prev := cfg.AfterConnect
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if prev != nil {
			if err := prev(ctx, conn); err != nil {
				return err
			}
		}
		return pgxvec.RegisterTypes(ctx, conn)
	}
}
