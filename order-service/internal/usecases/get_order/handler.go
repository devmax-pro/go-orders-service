package get_order

import (
	"context"
	"fmt"
	"github.com/devmax-pro/order-service/internal/adapters/cache"
	"github.com/devmax-pro/order-service/internal/entities"
)

type Handler struct {
	repo entities.OrderRepository
	csh  cache.Cache[entities.Order]
}

func New(repo entities.OrderRepository, csh cache.Cache[entities.Order]) *Handler {
	return &Handler{repo, csh}
}

func (h *Handler) Handle(ctx context.Context, id string) (*entities.Order, error) {
	cachedOrder, ex := h.csh.Get(id)
	if ex {
		return &cachedOrder, nil
	}

	order, err := h.repo.GetOrderById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order not found error: %w", err)
	}

	err = h.csh.Set(order.OrderUID, *order)
	if err != nil {
		return nil, fmt.Errorf("order cached failed: %w", err)
	}

	return order, nil
}
