package service

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
)

func (s *OrderServiceServer) GetOrderByID(ctx context.Context, req *orderpb.GetOrderByIDRequest) (*orderpb.GetOrderByIDResponse, error) {
	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required and must be positive")
	}

	order, err := s.service.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	respOrder := &orderpb.Order{
		OrderId:        order.OrderID,
		UserId:         order.UserID,
		ExpirationTime: timestamppb.New(order.ExpirationTime),
		Status:         domain.GetStringFromStatus(order.Status),
		Weight:         int32(order.Weight),
		Cost:           int32(order.Cost),
	}

	return &orderpb.GetOrderByIDResponse{Order: respOrder}, nil
}
