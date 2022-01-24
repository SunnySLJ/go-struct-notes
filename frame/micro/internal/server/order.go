package server

import (
	"context"
	"go-micro/gen/proto/order"
	"go-micro/internal/model"
	"go-micro/internal/service"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListOrders(ctx context.Context, req *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	orders, count, err := service.NewOrderService().List(ctx, int(req.PageNumber), int(req.PageSize), nil)
	if err != nil {
		return nil, err
	}
	resp := new(order.ListOrdersResponse)
	resp.Count = int32(count)
	resp.Orders = make([]*order.Order, len(orders))
	for k, v := range orders {
		resp.Orders[k] = &order.Order{
			Id:         v.Id,
			Name:       v.Name,
			Price:      float32(v.Price),
			CreateTime: timestamppb.New(v.CreateTime),
			UpdateTime: timestamppb.New(v.UpdateTime),
		}
	}
	return resp, nil
}

func (s *Server) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	mOrder := &model.Order{
		Id:    req.Order.Id,
		Name:  req.Order.Name,
		Price: float64(req.Order.Price),
	}
	err := service.NewOrderService().Create(ctx, mOrder)
	if err != nil {
		return nil, err
	}

	return &order.CreateOrderResponse{
		Order: &order.Order{
			Id:         mOrder.Id,
			Name:       mOrder.Name,
			Price:      float32(mOrder.Price),
			CreateTime: timestamppb.New(mOrder.CreateTime),
			UpdateTime: timestamppb.New(mOrder.UpdateTime),
		},
	}, nil
}

func (s *Server) UpdateOrder(ctx context.Context, req *order.UpdateOrderRequest) (*emptypb.Empty, error) {
	updateOrder := &model.Order{
		Name:  req.Order.Name,
		Price: float64(req.Order.Price),
	}
	updated := model.NewOrderOptionsRawString(updateOrder, req.UpdateMask.Paths...)

	condOrder := &model.Order{
		Id: req.Order.Id,
	}
	condition := model.NewOrderOptions(condOrder, model.OrderFieldId)

	err := service.NewOrderService().Update(ctx, updated, condition)
	return &emptypb.Empty{}, err
}

func (s *Server) GetOrder(ctx context.Context, req *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	condOrder := &model.Order{
		Name: req.Name,
	}
	condition := model.NewOrderOptions(condOrder, model.OrderFieldName)

	orders, _, err := service.NewOrderService().List(ctx, 0, 1, condition)
	if err != nil {
		return nil, err
	} else if len(orders) == 0 {
		return nil, errors.New("no order matched")
	}
	return &order.GetOrderResponse{
		Order: &order.Order{
			Id:         orders[0].Id,
			Name:       orders[0].Name,
			Price:      float32(orders[0].Price),
			CreateTime: timestamppb.New(orders[0].CreateTime),
			UpdateTime: timestamppb.New(orders[0].UpdateTime),
		},
	}, nil
}

func (s *Server) DeleteOrder(ctx context.Context, req *order.DeleteOrderRequest) (*emptypb.Empty, error) {
	condOrder := &model.Order{
		Name: req.Name,
	}
	condition := model.NewOrderOptions(condOrder, model.OrderFieldName)

	return &emptypb.Empty{}, service.NewOrderService().Delete(ctx, condition)
}
