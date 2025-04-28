package tx_manager

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
)

type TransactionManager interface {
	GetQueryEngine(ctx context.Context) db.DB
	RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error
	RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error
}
