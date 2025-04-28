package tx_manager

import (
	"context"
	"github.com/jackc/pgx/v4"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
)

type txManagerKey struct{}

type TxManager struct {
	db db.DB
}

func NewTxManager(db db.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}

	return m.beginFunc(ctx, opts, fn)
}

func (m *TxManager) RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}

	return m.beginFunc(ctx, opts, fn)
}

func (m *TxManager) RunRepeatableRead(ctx context.Context, fn func(ctxTx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}

	return m.beginFunc(ctx, opts, fn)
}

func (m *TxManager) beginFunc(ctx context.Context, opts pgx.TxOptions, fn func(ctxTx context.Context) error) error {
	tx, err := m.db.GetPool().BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = context.WithValue(ctx, txManagerKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (m *TxManager) GetQueryEngine(ctx context.Context) db.DB {
	v, ok := ctx.Value(txManagerKey{}).(db.DB)
	if ok && v != nil {
		return v
	}

	return m.db
}
