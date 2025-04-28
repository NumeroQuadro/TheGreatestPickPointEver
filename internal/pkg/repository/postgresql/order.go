package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/cache"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/logger"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
	"go.uber.org/zap"
	"log"
	"time"
)

type OrderRepo struct {
	tx     *tx_manager.TxManager
	client cache.Client
}

func NewOrdersRepo(tx *tx_manager.TxManager, client *cache.Client) *OrderRepo {
	return &OrderRepo{
		tx:     tx,
		client: *client,
	}
}

func (o *OrderRepo) Create(
	ctx context.Context,
	orderID int64,
	userID int64,
	expirationDate time.Time,
	weight int,
	cost int,
) (int64, error) {
	var (
		id int64
	)
	cacheKey := fmt.Sprintf("order_%d", orderID)

	value, err := o.client.GetOrdersFromCache(cacheKey)
	if err == nil && len(value) == 1 {
		return value[0].UserID, nil
	}

	err = o.tx.GetQueryEngine(ctx).ExecQueryRow(ctx,
		`
		INSERT INTO orders(order_id, user_id, expiration_date, weight, cost) 
		VALUES ($1,$2,$3, $4,$5) returning order_id;`,
		orderID,
		userID,
		expirationDate,
		weight,
		cost).Scan(&id)
	if err != nil {
		return 0, err
	}

	_, err = o.Find(ctx, orderID) // call Find to store value in cache
	if err != nil {
		return 0, err
	}

	return id, err
}

func (o *OrderRepo) Find(ctx context.Context, orderID int64) (domain.Order, error) {
	cacheKey := fmt.Sprintf("order_%d", orderID)
	value, err := o.client.GetOrdersFromCache(cacheKey)
	if err == nil && len(value) == 1 {
		logger.ZapLogger.Debug("returned from cache")

		return value[0], nil
	}

	order := domain.Order{}
	err = o.tx.GetQueryEngine(ctx).Get(ctx, &order, `SELECT 
		order_id,
		user_id,
		expiration_date,
		status,
		weight,
		cost,
		last_changed_at
	FROM orders
	WHERE order_id = $1;
	`, orderID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}

		return domain.Order{}, err
	}

	orders := []domain.Order{order}
	if err := o.client.SetOrdersToCache(cacheKey, orders); err == nil {
		log.Print("cache successfully updated")
	}

	return order, nil
}

func (o *OrderRepo) FindAll(ctx context.Context, filter repository.Filter, lastId *int64, limit *int) ([]domain.Order, error) {
	var orders []domain.Order
	var err error

	if lastId == nil || limit == nil {
		err = o.findAllWithoutPagination(ctx, filter, &orders)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrOrderNotFound
			}

			return nil, err
		}
	}
	if lastId != nil && limit != nil {
		err = o.findAllWithPagination(ctx, filter, *lastId, *limit, &orders)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrOrderNotFound
			}

			return nil, err
		}
	}

	return orders, nil
}

func (o *OrderRepo) Update(
	ctx context.Context,
	orderID int64,
	userID int64,
	expirationDate time.Time,
	status domain.Status,
	weight int,
	cost int,
) (int64, error) {
	var returnedOrderID int64
	cacheKey := fmt.Sprintf("order_%d", orderID)

	o.client.InvalidateOrderCache(cacheKey)
	err := o.tx.GetQueryEngine(ctx).ExecQueryRow(ctx,
		`
	UPDATE orders SET 
				  user_id = $2, 
				  expiration_date = $3, 
				  status = $4, 
				  weight = $5, 
				  cost = $6, 
				  last_changed_at = $7 
	            WHERE order_id = $1 
	RETURNING order_id;`,
		orderID,
		userID,
		expirationDate,
		status,
		weight,
		cost,
		time.Now()).Scan(&returnedOrderID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrOrderNotFound
		}
		logger.ZapLogger.Error("failed to process a request", zap.String("orderrepo", err.Error()))

		return 0, err
	}

	value, err := o.Find(ctx, returnedOrderID)
	if err != nil {
		return 0, err
	}

	return value.OrderID, nil
}

func (o *OrderRepo) Delete(ctx context.Context, orderID int64) error {
	cacheKey := fmt.Sprintf("order_%d", orderID)

	o.client.InvalidateOrderCache(cacheKey)
	execResult, err := o.tx.GetQueryEngine(ctx).Exec(ctx, `DELETE FROM orders WHERE order_id = $1;`, orderID)
	if err != nil {
		return err
	}

	if execResult.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (o *OrderRepo) findAllWithPagination(
	ctx context.Context,
	filter repository.Filter,
	lastID int64,
	limit int,
	orders *[]domain.Order,
) error {
	cacheKey := o.buildFindAllCacheKey(filter, &lastID, &limit)
	cacheOrders, err := o.client.GetOrdersFromCache(cacheKey)
	if err == nil {
		*orders = cacheOrders
		logger.ZapLogger.Debug("return from cache")

		return nil
	}

	baseQuery, values := repository.BuildSQLQuery(filter)
	values = append(values, lastID, limit)

	baseQuery += fmt.Sprintf(" AND order_id > $%d LIMIT $%d", len(values)-1, len(values))
	err = o.tx.GetQueryEngine(ctx).Select(ctx, orders, baseQuery, values...)

	if err := o.client.SetOrdersToCache(cacheKey, *orders); err != nil {
		logger.ZapLogger.Error("failed to cache orders", zap.String("orderrepo", err.Error()))
	}

	return err
}

func (o *OrderRepo) findAllWithoutPagination(
	ctx context.Context,
	filter repository.Filter,
	orders *[]domain.Order,
) error {
	cacheKey := o.buildFindAllCacheKey(filter, nil, nil)
	cacheOrders, err := o.client.GetOrdersFromCache(cacheKey)
	if err == nil {
		*orders = cacheOrders
		logger.ZapLogger.Debug("return from cache")

		return nil
	}

	baseQuery, values := repository.BuildSQLQuery(filter)
	err = o.tx.GetQueryEngine(ctx).Select(ctx, orders, baseQuery, values...)

	if err := o.client.SetOrdersToCache(cacheKey, *orders); err != nil {
		logger.ZapLogger.Error("failed to cache orders", zap.String("orderrepo", err.Error()))
	}

	return err
}

func (o *OrderRepo) buildFindAllCacheKey(filter repository.Filter, lastID *int64, limit *int) string {
	filterString := filter.GetFilterStringView()
	base := fmt.Sprintf("findAll:%v", filterString)

	if lastID != nil && limit != nil {
		base = fmt.Sprintf("%v,lastId,limit", base)
	}

	return base
}
