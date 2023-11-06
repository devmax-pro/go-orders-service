package app

import (
	"context"
	"errors"
	"github.com/devmax-pro/order-service/internal/adapters/cache"
	"github.com/devmax-pro/order-service/internal/adapters/storage/postgres"
	"github.com/devmax-pro/order-service/internal/entities"
	"github.com/devmax-pro/order-service/internal/services/cache_restorer"
	"github.com/devmax-pro/order-service/internal/services/graceful_shutdown"
	"net/http"
	"os"
	"time"

	"github.com/devmax-pro/order-service/internal/adapters/http/router"
	"github.com/devmax-pro/order-service/internal/adapters/subscriber"
	"github.com/devmax-pro/order-service/internal/usecases/add_order"

	"github.com/devmax-pro/order-service/internal/adapters/http/controller"
	"github.com/devmax-pro/order-service/internal/adapters/http/server"
	logs "github.com/devmax-pro/order-service/internal/adapters/logger"
	"github.com/devmax-pro/order-service/internal/usecases/get_order"
)

func Run() {
	if err := logs.Initialize("debug"); err != nil {
		return
	}

	DbSource := os.Getenv("DB_SOURCE") // @TODO parse url from .env file
	db, err := postgres.New(DbSource)
	if err != nil {
		logs.Error("Error occurred while init postgres db", err)
	}

	err = db.Ping()
	if err != nil {
		logs.Fatal("Error ping database", err)
	}

	repo := postgres.NewOrders(db)
	if err != nil {
		logs.Fatal("Error occurred while init subscriber", err)
	}

	csh := cache.NewMemoryCache[entities.Order]()

	rst := cache_restorer.New(db, csh)
	err = rst.RestoreOrders()
	if err != nil {
		logs.Fatal("Error occurred while restore cache", err)
	}

	getOrderHandler := get_order.New(repo, csh)
	ctrl := controller.New(getOrderHandler)
	rtr := router.New(ctrl)
	srv := server.New(rtr)

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logs.Error("Error occurred while running http server", err)
		}
	}()

	NatsURL := os.Getenv("NATS_URL")
	sb, err := subscriber.New(NatsURL)
	if err != nil {
		logs.Fatal("Error occurred while init subscriber", err)
	}

	addOrderHandler := add_order.New(repo, csh)
	err = sb.Subscribe("orders-channel", addOrderHandler)

	if err != nil {
		logs.Fatal("Error occurred while subscriber trying subscribe to channel", err)
	}
	logs.Info("Initialization App is successful")

	// wait for termination signal and register all clean-up operations
	wait := graceful_shutdown.GracefulShutdown(context.Background(), 10*time.Second, map[string]graceful_shutdown.Operation{
		"database": func(ctx context.Context) error {
			db.Close()
			return nil
		},
		"http-server": func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
		"subscriber": func(ctx context.Context) error {
			return sb.Close()
		},
	})

	<-wait
}
