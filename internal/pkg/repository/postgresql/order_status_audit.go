package postgresql

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
)

type OrderStatusAuditRepository interface {
	Create(ctx context.Context, job domain.AuditOrderInfo) (int64, error)
}

type OrderStatusAuditRepositoryImpl struct {
	db db.DB
}

func NewOrderStatusAuditRepositoryImpl(database db.DB) *OrderStatusAuditRepositoryImpl {
	return &OrderStatusAuditRepositoryImpl{db: database}
}

func (a *OrderStatusAuditRepositoryImpl) Create(ctx context.Context, job domain.AuditOrderInfo) (int64, error) {
	var entryID int64

	query := `
		INSERT INTO order_status_audit (
			order_id, previous_status, current_status
		) VALUES (
			$1, $2, $3
		) returning entry_id;
	`

	err := a.db.ExecQueryRow(ctx,
		query,
		job.OrderID,
		job.PreviousStatus,
		job.CurrentStatus,
	).Scan(&entryID)

	if err != nil {
		return 0, err
	}

	return entryID, nil
}
