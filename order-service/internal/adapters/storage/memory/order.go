package memory

import (
	"context"
	"errors"
	"github.com/devmax-pro/order-service/internal/entities"
)

type OrderRepository struct {
	orders map[string]*entities.Order
}

func NewOrders(orders map[string]*entities.Order) *OrderRepository {
	return &OrderRepository{orders}
}

func (repo *OrderRepository) GetOrderById(ctx context.Context, id string) (*entities.Order, error) {
	order, ex := repo.orders[id]
	if ex {
		return order, nil
	}
	return nil, errors.New("order not found")
}

func (repo *OrderRepository) AddOrder(ctx context.Context, order *entities.Order) (err error) {
	repo.orders[order.OrderUID] = order
	return nil
}
