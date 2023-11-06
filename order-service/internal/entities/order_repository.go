package entities

import "context"

type OrderRepository interface {
	GetOrderById(ctx context.Context, id string) (*Order, error)
	AddOrder(ctx context.Context, order *Order) error
}
