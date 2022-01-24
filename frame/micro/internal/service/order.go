package service

import (
	"context"
	"go-micro/internal/dao"
	"go-micro/internal/logx"
	"go-micro/internal/metrics"
	"go-micro/internal/model"
	"go-micro/internal/mysql"

	"github.com/pkg/errors"
)

type OrderService struct {
	orderRepo model.OrderModel
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo: dao.NewOrderRepo(mysql.DB),
	}
}

func (orderSvc *OrderService) List(ctx context.Context, pageNumber, pageSize int, condition *model.OrderOptions) ([]model.Order, int64, error) {
	metrics.OrderList.With(map[string]string{"service": "example"}).Inc()
	logx.WithTrace(ctx).Infof("page number is %d", pageNumber)

	orders, err := orderSvc.orderRepo.QueryOrders(ctx, pageNumber, pageSize, condition)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "OrderService List pageNumber %d pageSize %d condition %+v", pageNumber, pageSize, condition)
	}
	count, err := orderSvc.orderRepo.CountOrders(ctx, condition)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "OrderService Count condition %+v", condition)
	}

	return orders, count, nil
}

func (orderSvc *OrderService) Create(ctx context.Context, order *model.Order) error {
	err := orderSvc.orderRepo.AddOrder(ctx, order)
	if err != nil {
		return errors.Wrapf(err, "OrderService Create  order %+v", order)
	}
	return nil
}

func (orderSvc *OrderService) Update(ctx context.Context, updated, condition *model.OrderOptions) error {
	err := orderSvc.orderRepo.UpdateOrder(ctx, updated, condition)
	if err != nil {
		return errors.Wrapf(err, "OrderService Update updated %+v condition %+v", updated, condition)
	}
	return nil
}

func (orderSvc *OrderService) Delete(ctx context.Context, condition *model.OrderOptions) error {
	err := orderSvc.orderRepo.DeleteOrder(ctx, condition)
	if err != nil {
		return errors.Wrapf(err, "OrderService Delete condition %+v", condition)
	}
	return nil
}
