package cache_restorer

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/devmax-pro/order-service/internal/adapters/cache"
	"github.com/devmax-pro/order-service/internal/adapters/storage/postgres"
	"github.com/devmax-pro/order-service/internal/entities"
)

type CacheRestorer struct {
	db  *postgres.Postgres
	csh cache.Cache[entities.Order]
}

func New(db *postgres.Postgres, csh cache.Cache[entities.Order]) *CacheRestorer {
	return &CacheRestorer{db, csh}
}

func (cr *CacheRestorer) RestoreOrders() error {

	sql, args, err := cr.db.Builder.
		Select("o.order_uid", "o.track_number", "o.entry", "o.order_locale", "o.internal_signature", "o.customer_id",
			"o.delivery_service", "o.shard_key", "o.sm_id", "o.date_created", "o.oof_shard",
			// select delivery
			"d.delivery_name", "d.phone", "d.zip", "d.city", "d.address", "d.region", "d.email",
			// select payment
			"p.payment_transaction", "p.request_id", "p.currency", "p.payment_provider", "p.amount",
			"p.payment_dt", "p.bank", "p.delivery_cost", "p.goods_total", "p.custom_fee").
		From("orders as o").
		Join("order_deliveries as d on o.delivery_id = d.id").
		Join("order_payments as p on o.payment_id = p.id").
		ToSql()
	if err != nil {
		return fmt.Errorf("error building sql for query of order: %w", err)
	}

	rows, err := cr.db.Pool.Query(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("error query order rows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		order := entities.Order{}
		delivery := entities.Delivery{}
		payment := entities.Payment{}
		err = rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
			// scan delivery
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region,
			&delivery.Email,
			// scan payment
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
		)
		if err != nil {
			return fmt.Errorf("error scanning order row: %w", err)
		}

		// order items
		sql, args, err = cr.db.Builder.
			Select("chrt_id", "track_number", "price", "rid", "item_name", "sale", "size", "total_price", "nm_id", "brand", "status").
			From("order_items").
			Where(squirrel.Eq{"order_uid": order.OrderUID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("error building sql for query of order items: %w", err)
		}

		itemRows, err := cr.db.Pool.Query(context.Background(), sql, args...)
		if err != nil {
			return fmt.Errorf("error query order items rows: %w", err)
		}
		defer itemRows.Close()

		var orderItems []entities.Item
		for itemRows.Next() {
			var item entities.Item
			err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
				&item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				return fmt.Errorf("unable to scan order_item row: %w", err)
			}
			orderItems = append(orderItems, item)
		}

		if itemRows.Err() != nil {
			return fmt.Errorf("error occurred during rows processing: %w", rows.Err())
		}

		order.Items = orderItems

		err = cr.csh.Set(order.OrderUID, order)
		if err != nil {
			return fmt.Errorf("order cached failed: %w", err)
		}
	}

	if rows.Err() != nil {
		return fmt.Errorf("error occurred during rows processing: %w", rows.Err())
	}
	return nil
}
