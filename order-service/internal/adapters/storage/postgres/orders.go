package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/devmax-pro/order-service/internal/entities"
	"github.com/jackc/pgx/v5"
)

type OrderRepository struct {
	db *Postgres
}

func NewOrders(db *Postgres) *OrderRepository {
	return &OrderRepository{db}
}

func (repo *OrderRepository) GetOrderById(ctx context.Context, id string) (*entities.Order, error) {

	sql, args, err := repo.db.Builder.
		Select("o.order_uid", "o.track_number", "o.entry", "o.order_locale", "o.internal_signature", "o.customer_id",
			"o.delivery_service", "o.shard_key", "o.sm_id", "o.date_created", "o.oof_shard", "delivery_id",
			// select delivery
			"d.delivery_name", "d.phone", "d.zip", "d.city", "d.address", "d.region", "d.email",
			// select payment
			"p.payment_transaction", "p.request_id", "p.currency", "p.payment_provider", "p.amount",
			"p.payment_dt", "p.bank", "p.delivery_cost", "p.goods_total", "p.custom_fee").
		From("orders as o").
		Join("order_deliveries as d on o.delivery_id = d.id").
		Join("order_payments as p on o.payment_id = p.id").
		Where(squirrel.Eq{"order_uid": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building sql for query of order: %w", err)
	}

	row := repo.db.Pool.QueryRow(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error query row for order: %w", err)
	}

	order := entities.Order{}
	delivery := entities.Delivery{}
	payment := entities.Payment{}
	err = row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
		// scan delivery
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
		&delivery.Region, &delivery.Email,
		// scan payment
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	// order items
	sql, args, err = repo.db.Builder.
		Select("chrt_id", "track_number", "price", "rid", "item_name", "sale", "size", "total_price", "nm_id", "brand", "status").
		From("order_items").
		Where(squirrel.Eq{"order_uid": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building sql for query of order items: %w", err)
	}

	rows, err := repo.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error query order items rows: %w", err)
	}
	defer rows.Close()

	var orderItems []entities.Item
	for rows.Next() {
		var item entities.Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		orderItems = append(orderItems, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error occurred during rows processing: %w", rows.Err())
	}

	order.Items = orderItems

	return &order, nil
}

func (repo *OrderRepository) AddOrder(ctx context.Context, order *entities.Order) (err error) {

	return repo.db.Transact(func(tx pgx.Tx) error {
		// Save delivery
		delivery := order.Delivery
		sql, args, err := repo.db.Builder.
			Insert("order_deliveries").
			Columns("delivery_name", "phone", "zip", "city", "address", "region", "email").
			Values(delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address,
				delivery.Region, delivery.Email).
			Suffix("RETURNING \"id\"").
			ToSql()
		if err != nil {
			return fmt.Errorf("error building delivery sql: %w", err)
		}

		var deliveryID uint64
		err = tx.QueryRow(ctx, sql, args...).Scan(&deliveryID)
		if err != nil {
			return fmt.Errorf("error executing order delivery insert sql: %w", err)
		}

		// Save payment
		payment := order.Payment
		sql, args, err = repo.db.Builder.
			Insert("order_payments").
			Columns("payment_transaction", "request_id", "currency", "payment_provider", "amount",
				"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
			Values(payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount,
				payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee).
			Suffix("RETURNING \"id\"").
			ToSql()
		if err != nil {
			return fmt.Errorf("error building payment sql: %w", err)
		}

		var paymentID uint64
		err = tx.QueryRow(ctx, sql, args...).Scan(&paymentID)
		if err != nil {
			return fmt.Errorf("error executing order payment insert sql: %w", err)
		}

		// Save order
		sql, args, err = repo.db.Builder.
			Insert("orders").
			Columns("order_uid", "track_number", "entry", "order_locale", "internal_signature", "customer_id",
				"delivery_service", "shard_key", "sm_id", "date_created", "oof_shard", "delivery_id").
			Values(order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
				order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated,
				order.OofShard, deliveryID).
			ToSql()
		if err != nil {
			return fmt.Errorf("error building sql: %w", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("error executing insert order sql: %w", err)
		}

		// Save order items
		for _, item := range order.Items {
			sql, args, err = repo.db.Builder.
				Insert("order_items").
				Columns("order_uid", "chrt_id", "track_number", "price", "rid", "item_name", "sale", "size", "total_price", "nm_id", "brand", "status").
				Values(order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size,
					item.TotalPrice, item.NmID, item.Brand, item.Status).
				ToSql()
			if err != nil {
				return fmt.Errorf("error building order item sql: %w", err)
			}

			_, err = tx.Exec(ctx, sql, args...)
			if err != nil {
				return fmt.Errorf("error executing order item insert sql: %w", err)
			}
		}

		return nil
	})
}
