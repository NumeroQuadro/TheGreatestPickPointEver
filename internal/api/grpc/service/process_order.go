package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
)

func (s *OrderServiceServer) ProcessOrder(ctx context.Context, req *orderpb.ProcessOrderRequest) (*orderpb.ProcessOrderResponse, error) {
	switch req.GetAction() {
	case "complete":
		err := s.service.CompleteOrder(ctx, req.GetOrderId(), req.GetUserId())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	case "refund":
		err := s.service.RefundOrder(ctx, req.GetOrderId(), s.config.OrderExpirationDays)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Invalid action")
	}

	return &orderpb.ProcessOrderResponse{}, nil
}
