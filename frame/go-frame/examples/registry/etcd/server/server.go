package server

import (
	"context"
	"go-frame/examples/proto/order"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	// 使用unsafe可以强制让编译器检查是否实现了相关方法
	order.UnsafeOrderServiceServer
}

func (s *Server) ListOrders(ctx context.Context, in *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	return &order.ListOrdersResponse{Count: 100}, nil
}

func (s *Server) CreateOrder(ctx context.Context, request *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	panic("implement me")
}

func (s *Server) UpdateOrder(ctx context.Context, request *order.UpdateOrderRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

func (s *Server) GetOrder(ctx context.Context, request *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	panic("implement me")
}

func (s *Server) DeleteOrder(ctx context.Context, request *order.DeleteOrderRequest) (*emptypb.Empty, error) {
	panic("implement me")
}
