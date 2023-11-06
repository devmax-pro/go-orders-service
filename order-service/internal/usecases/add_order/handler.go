package add_order

import (
	"context"
	"encoding/json"
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

func (h *Handler) Handle(ctx context.Context, data []byte) error {
	order := entities.Order{}
	err := json.Unmarshal(data, &order)
	if err != nil {
		return fmt.Errorf("message unmarshal failed: %w", err)
	}
	if order.OrderUID == "" {
		return fmt.Errorf("order uid must not be empty")
	}

	err = h.csh.Set(order.OrderUID, order)
	if err != nil {
		return fmt.Errorf("order cached failed: %w", err)
	}

	err = h.repo.AddOrder(ctx, &order)
	if err != nil {
		return fmt.Errorf("order saving to database failed: %w", err)
	}

	return nil
}
