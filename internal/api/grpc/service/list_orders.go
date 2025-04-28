package service

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
)

func (s *OrderServiceServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	if req.GetStatus() != "" {
		if !domain.IsStatusValid(req.GetStatus()) {
			return nil, status.Error(codes.InvalidArgument, "status should not be empty")
		}
	}

	if req.GetUserId() > 0 {
		var limitVal *int
		var lastIDVal *int64
		if req.GetLimit() != 0 {
			l := int(req.GetLimit())
			limitVal = &l
		}
		if req.GetLastId() > 0 {
			lastIDVal = &req.LastId
		}
		orders, err := s.service.GetOrdersByUserID(ctx, req.GetUserId(), limitVal, lastIDVal)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return convertOrdersResponse(orders), nil
	}

	searchFilter := &service.SearchFilter{}
	if req.GetSearchTerm() != "" {
		term := req.GetSearchTerm()
		searchFilter = &service.SearchFilter{SearchTerm: &term}
	}

	if req.GetStatus() != "" {
		st := req.GetStatus()
		searchFilter.Status = &st
	}

	if req.GetUserId() > 0 {
		userID := req.GetUserId()
		searchFilter.UserID = &userID
	}

	var limitVal *int
	var lastIDVal *int64
	if req.GetLimit() > 0 {
		l := int(req.GetLimit())
		limitVal = &l
	}
	if req.GetLastId() > 0 {
		lastIDVal = &req.LastId
	}

	orders := s.service.GetOrders(ctx, lastIDVal, limitVal, searchFilter)
	if len(orders) == 0 {
		log.Println("ListOrders: no orders found")
	}

	return convertOrdersResponse(orders), nil
}

func convertOrdersResponse(orders []domain.Order) *orderpb.ListOrdersResponse {
	respOrders := make([]*orderpb.Order, len(orders))
	for i, o := range orders {
		respOrders[i] = &orderpb.Order{
			OrderId:        o.OrderID,
			UserId:         o.UserID,
			ExpirationTime: timestamppb.New(o.ExpirationTime),
			Status:         domain.GetStringFromStatus(o.Status),
			Weight:         int32(o.Weight),
			Cost:           int32(o.Cost),
		}
	}

	return &orderpb.ListOrdersResponse{Orders: respOrders}
}
