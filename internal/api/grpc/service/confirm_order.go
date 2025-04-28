package service

import (
	"context"
	"errors"
	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderServiceServer) ConfirmOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if !s.isOrderRequestValid(req) {
		return nil, status.Error(codes.InvalidArgument, "order is not valid")
	}

	packageTypeFromString, err := domain.GetPackageTypeFromString(req.GetPackageType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unable to retrieve order package type")
	}

	dto := service.OrderDto{
		OrderID:        req.GetOrderId(),
		UserID:         req.GetUserId(),
		ExpirationTime: req.GetExpirationTime().AsTime(),
		Weight:         int(req.GetWeight()),
		Cost:           int(req.GetCost()),
	}

	order, err := s.service.AddOrder(ctx, dto, packageTypeFromString, req.GetIsAdditionalFilm())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderAlreadyExists):
			return nil, status.Error(codes.Internal, "order already exists")
		default:
			return nil, status.Error(codes.Internal, "unable to create an order")
		}
	}

	return &orderpb.CreateOrderResponse{OrderId: order.OrderID}, nil
}
