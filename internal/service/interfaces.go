//go:generate mockgen -source ./interfaces.go -destination=./mocks/interfaces.go -package=mock_repository
package service

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository"
	"time"
)

type OrderRepository interface {
	Create(
		ctx context.Context,
		orderID int64,
		userID int64,
		expirationDate time.Time,
		weight int,
		cost int) (int64, error)
	Find(
		ctx context.Context,
		orderID int64,
	) (domain.Order, error)
	FindAll(
		ctx context.Context,
		filter repository.Filter,
		lastID *int64,
		limit *int,
	) ([]domain.Order, error)
	Update(
		ctx context.Context,
		orderID int64,
		userID int64,
		expirationDate time.Time,
		status domain.Status,
		weight int,
		cost int,
	) (int64, error)
	Delete(
		ctx context.Context,
		orderID int64,
	) error
}

type AuditEntriesRepository interface {
	Create(ctx context.Context)
}
