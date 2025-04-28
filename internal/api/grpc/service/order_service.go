package service

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"

	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
)

type OrderService interface {
	AddOrder(ctx context.Context,
		orderDto service.OrderDto,
		packageType domain.PackageType,
		isAdditionalFilm bool) (domain.Order, error)
	RetrieveOrdersFromFile(ctx context.Context,
		data []byte) error
	CompleteOrder(ctx context.Context,
		orderID int64,
		userID int64) error
	ReturnOrder(ctx context.Context,
		orderID int64) error
	RefundOrder(ctx context.Context,
		orderID int64,
		expirationDays int) error
	GetOrdersByUserID(ctx context.Context,
		userID int64,
		limit *int,
		lastID *int64) ([]domain.Order, error)
	GetOrderByID(ctx context.Context,
		orderID int64) (domain.Order, error)
	GetOrders(ctx context.Context,
		lastID *int64,
		limit *int,
		searchFilter *service.SearchFilter) []domain.Order
	GetOrdersBySpecificStatus(ctx context.Context,
		status domain.Status) []domain.Order
	GetRefundedOrders(ctx context.Context,
		lastID *int64,
		limit *int) ([]domain.Order, error)
}

type OrderServiceServer struct {
	orderpb.UnimplementedOrderServiceServer
	service OrderService
	config  config.Config
}

func NewOrderServiceServer(service OrderService, config config.Config) *OrderServiceServer {
	return &OrderServiceServer{service: service, config: config}
}

func (s *OrderServiceServer) isOrderRequestValid(req *orderpb.CreateOrderRequest) bool {
	if req.GetUserId() <= 0 {
		return false
	}
	if req.GetCost() <= 0 {
		return false
	}
	if req.GetWeight() <= 0 {
		return false
	}

	return true
}
