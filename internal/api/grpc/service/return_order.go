package service

import (
	"context"
	"errors"

	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderServiceServer) ReturnOrder(ctx context.Context, req *orderpb.ReturnOrderRequest) (*orderpb.ReturnOrderResponse, error) {
	err := s.service.ReturnOrder(ctx, req.GetOrderId())
	if errors.Is(err, domain.ErrOrderNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orderpb.ReturnOrderResponse{}, nil
}
