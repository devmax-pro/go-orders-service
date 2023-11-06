package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

	logs "github.com/devmax-pro/order-service/internal/adapters/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnAttempts      = 5
	defaultMaxConns          = int32(4)
	defaultMinConns          = int32(0)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = time.Minute * 30
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = time.Second * 5
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
	Builder      squirrel.StatementBuilderType
	Pool         *pgxpool.Pool
}

func New(databaseUrl string) (*Postgres, error) {
	pg := &Postgres{
		connAttempts: defaultConnAttempts,
		Builder:      squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	poolConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		logs.Error("Parse config for postgres db error", err)
		return nil, err
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)
	poolConfig.MaxConns = defaultMaxConns
	poolConfig.MinConns = defaultMinConns
	poolConfig.MaxConnLifetime = defaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		logs.Error(fmt.Sprintf("Postgres is trying to connect, attempts left: %d", pg.connAttempts))

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		logs.Error("Postgres is trying to connect, attempts are over", err)
		return nil, err
	}
	logs.Info("Postgres is connected successful")

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func (p *Postgres) Ping() error {
	err := p.Pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	logs.Info("Ping of database is successful")
	return nil
}

func (p *Postgres) Transact(txFunc func(pgx.Tx) error) (err error) {
	tx, err := p.Pool.Begin(context.Background())
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(context.Background())
		} else {
			err = tx.Commit(context.Background())
		}
		return
	}()
	err = txFunc(tx)
	return err
}
