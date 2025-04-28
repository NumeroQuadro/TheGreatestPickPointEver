package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/monitoring"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
	"sort"
	"time"
)

type OrderServiceImpl struct {
	repo      OrderRepository
	txManager tx_manager.TxManager
	wm        *workers.WorkerManager
}

func NewOrderServiceImpl(repo OrderRepository, txManager tx_manager.TxManager, wm *workers.WorkerManager) *OrderServiceImpl {
	return &OrderServiceImpl{
		repo:      repo,
		txManager: txManager,
		wm:        wm,
	}
}

func (o *OrderServiceImpl) AddOrder(
	ctx context.Context,
	orderDto OrderDto,
	packageType domain.PackageType,
	isAdditionalFilm bool,
) (domain.Order, error) {
	or := ConvertDtoToDomainOrder(orderDto)
	if or.ExpirationTime.Before(time.Now()) {
		monitoring.OrdersFailedCreationTotal.Inc()

		return domain.Order{}, domain.ErrExpirationDateInPast
	}

	packagingStrategy, err := domain.GetPackagingStrategy(packageType)
	if err != nil {
		monitoring.OrdersFailedCreationTotal.Inc()

		return domain.Order{}, err
	}

	finalCost, err := applyPackagingStrategy(packagingStrategy, isAdditionalFilm, or.Weight, or.Cost)
	if err != nil {
		monitoring.OrdersFailedCreationTotal.Inc()

		return domain.Order{}, err
	}

	var order domain.Order
	if err := o.txManager.RunSerializable(ctx, func(_ context.Context) error {
		id, err := o.repo.Create(ctx, or.OrderID, or.UserID, or.ExpirationTime, or.Weight, finalCost)
		if err != nil {
			return err
		}
		order, err = o.repo.Find(ctx, id)

		return err
	}); err != nil {
		monitoring.OrdersFailedCreationTotal.Inc()

		return domain.Order{}, err
	}

	monitoring.OrdersCreatedTotal.Inc()

	return order, nil
}

func (o *OrderServiceImpl) RetrieveOrdersFromFile(ctx context.Context, data []byte) error {
	var newOrders []External
	if err := json.Unmarshal(data, &newOrders); err != nil {
		return fmt.Errorf("failed to unmarshal orders file: %w", err)
	}

	for _, order := range newOrders {
		_, err := o.repo.Create(ctx, order.OrderID, order.UserID, order.ExpirationTime, order.Weight, order.Cost)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *OrderServiceImpl) GetOrdersByUserID(
	ctx context.Context,
	userID int64,
	limit *int,
	lastID *int64,
) ([]domain.Order, error) {
	var (
		orders []domain.Order
		err    error
	)
	filter := repository.Filter{
		UserID:         &userID,
		ExpirationTime: nil,
		Status:         nil,
	}

	if err := o.txManager.RunReadUncommitted(ctx, func(_ context.Context) error {
		orders, err = o.repo.FindAll(ctx, filter, lastID, limit)

		return err
	}); err != nil {
		return []domain.Order{}, err
	}

	return orders, err
}

func (o *OrderServiceImpl) ReturnOrder(ctx context.Context, orderID int64) error {
	if err := o.txManager.RunSerializable(ctx, func(_ context.Context) error {
		or, err := o.repo.Find(ctx, orderID)
		if err != nil {
			return domain.ErrOrderNotFound
		}

		if or.Status == domain.Completed {
			return domain.ErrOrderAlreadyCompleted
		}

		if or.Status == domain.Confirmed {
			if or.ExpirationTime.After(time.Now()) {
				return domain.ErrExpirationDateInFuture
			}
		}

		if or.Status != domain.Refunded {
			return domain.ErrOrderHasToBeRefunded
		}

		return o.repo.Delete(ctx, orderID)
	}); err != nil {
		return fmt.Errorf("o.txManager.RunSerializable from ReturnOrder, err: %v", err)
	}
	monitoring.OrdersReturnedTotal.Inc()

	return nil
}

func (o *OrderServiceImpl) RefundOrder(ctx context.Context, orderID int64, expirationDays int) error {
	var (
		prevStatus domain.Status
		newStatus  domain.Status
	)

	if err := o.txManager.RunSerializable(ctx, func(_ context.Context) error {
		or, err := o.repo.Find(ctx, orderID)
		if err != nil {
			return domain.ErrOrderNotFound
		}

		if or.Status != domain.Completed {
			return domain.ErrOrderNotCompleted
		}

		refundPeriod := time.Duration(24*expirationDays) * time.Hour
		if or.LastChangedAt.After(time.Now().Add(refundPeriod)) {
			return domain.ErrOrderCannotBeRefunded
		}
		status := domain.Refunded
		prevStatus = or.Status

		_, err = o.repo.Update(ctx, or.OrderID, or.UserID, or.ExpirationTime, status, or.Weight, or.Cost)

		newStatus = status

		return err
	}); err != nil {
		return fmt.Errorf("o.txManager.RunSerializable from RefundOrder: %w", err)
	}

	info := domain.AuditOrderInfo{
		OrderID:        orderID,
		PreviousStatus: prevStatus,
		CurrentStatus:  newStatus,
	}
	o.wm.LogAudit(info)
	monitoring.OrdersRefundedTotal.Inc()

	return nil
}

func (o *OrderServiceImpl) GetOrderByID(ctx context.Context, orderID int64) (domain.Order, error) {
	var (
		order domain.Order
		err   error
	)
	if err := o.txManager.RunReadUncommitted(ctx, func(_ context.Context) error {
		order, err = o.repo.Find(ctx, orderID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (o *OrderServiceImpl) CompleteOrder(ctx context.Context, orderID int64, userID int64) error {
	var (
		prevStatus domain.Status
		newStatus  domain.Status
	)

	if err := o.txManager.RunSerializable(ctx, func(_ context.Context) error {
		or, err := o.repo.Find(ctx, orderID)
		if err != nil {
			return domain.ErrOrderNotFound
		}
		if or.UserID != userID {
			return domain.ErrOrderNotBelongToUser
		}
		if or.ExpirationTime.Before(time.Now()) {
			return domain.ErrExpirationDateInPast
		}

		if or.Status == domain.Completed {
			return domain.ErrOrderAlreadyCompleted
		}

		status := domain.Completed
		prevStatus = or.Status

		_, err = o.repo.Update(ctx, or.OrderID, or.UserID, or.ExpirationTime, status, or.Weight, or.Cost)

		newStatus = status

		return err
	}); err != nil {
		return fmt.Errorf("o.txManager.RunSerializable from CompleteOrder: %w", err)
	}

	info := domain.AuditOrderInfo{
		OrderID:        orderID,
		PreviousStatus: prevStatus,
		CurrentStatus:  newStatus,
	}
	o.wm.LogAudit(info)
	monitoring.OrdersCompletedTotal.Inc()

	return nil
}

func (o *OrderServiceImpl) GetRefundedOrders(ctx context.Context, lastID *int64, limit *int) ([]domain.Order, error) {
	var (
		orders []domain.Order
		err    error
	)
	status := domain.Refunded

	if err := o.txManager.RunRepeatableRead(ctx, func(_ context.Context) error {
		filter := repository.Filter{
			ExpirationTime: nil,
			Status:         &status,
			UserID:         nil,
		}
		orders, err = o.repo.FindAll(ctx, filter, lastID, limit)

		return err
	}); err != nil {
		return []domain.Order{}, nil
	}

	return orders, err
}

func (o *OrderServiceImpl) GetOrders(
	ctx context.Context,
	lastID *int64,
	limit *int,
	searchFilter *SearchFilter,
) []domain.Order {
	var (
		orders []domain.Order
		err    error
	)
	if err := o.txManager.RunRepeatableRead(ctx, func(_ context.Context) error {
		var filter repository.Filter
		if searchFilter != nil {
			filter = repository.Filter{
				OrderID:        searchFilter.OrderID,
				UserID:         searchFilter.UserID,
				ExpirationTime: searchFilter.ExpirationTime,
				Status:         (*domain.Status)(searchFilter.Status),
			}
			if searchFilter.SearchTerm != nil {
				term := "%" + *searchFilter.SearchTerm + "%"
				filter.SearchTerm = &term
			}
		}

		orders, err = o.repo.FindAll(ctx, filter, lastID, limit)
		if err != nil {
			return err
		}
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].LastChangedAt.Before(orders[j].LastChangedAt)
		})

		return nil
	}); err != nil {
		return []domain.Order{}
	}

	return orders
}

func (o *OrderServiceImpl) GetOrdersBySpecificStatus(ctx context.Context, status domain.Status) []domain.Order {
	var (
		orders []domain.Order
		err    error
	)

	if err := o.txManager.RunRepeatableRead(ctx, func(ctxTx context.Context) error {
		filter := repository.Filter{
			ExpirationTime: nil,
			Status:         &status,
			UserID:         nil,
		}

		orders, err = o.repo.FindAll(ctx, filter, nil, nil)
		if err != nil {
			return fmt.Errorf("o.repository.FindAll: %w", err)
		}

		return nil
	}); err != nil {
		return []domain.Order{}
	}

	return orders
}

func applyPackagingStrategy(
	packageStrategy domain.Package,
	withAdditionalFilm bool,
	weight int,
	cost int,
) (int, error) {
	if err := packageStrategy.ValidatePackagedOrder(weight); err != nil {
		return 0, err
	}

	totalPackagingCost := packageStrategy.GetCost()

	if withAdditionalFilm {
		filmPackage := domain.FilmPackage{}

		if err := filmPackage.ValidatePackagedOrder(weight); err != nil {
			return 0, err
		}

		totalPackagingCost += filmPackage.GetCost()
	}

	return cost + totalPackagingCost, nil
}
