package controller

import (
	"encoding/json"
	"github.com/devmax-pro/order-service/internal/entities"
	"github.com/devmax-pro/order-service/internal/usecases/get_order"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Controller struct {
	GetOrderHandler *get_order.Handler
}

func New(h *get_order.Handler) *Controller {
	return &Controller{h}
}

func (c Controller) GetOrder(w http.ResponseWriter, r *http.Request) {

	var order *entities.Order
	var err error

	if orderID := chi.URLParam(r, "orderId"); orderID != "" {
		order, err = c.GetOrderHandler.Handle(r.Context(), orderID)
		if err != nil {
			NotFound("Finding order from database error", err, w, r)
			return
		}
	} else {
		BadRequest("Extract order id from path error", err, w, r)
		return
	}

	payload, err := json.Marshal(order)
	if err != nil {
		InternalError("Marshal order to payload error", err, w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(payload)
	if err != nil {
		InternalError("Write payload to response error", err, w, r)
		return
	}
}
