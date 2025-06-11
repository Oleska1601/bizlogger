package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	_defaultMaxPoolSize     = 1
	_defaultMaxConnAttempts = 10
	_defaultMaxConnTimeout  = time.Second
)

// инфраструктурный слой, отвечающий за общие настройки подключения и управлением пулом соединений
type Postgres struct {
	maxPoolSize     int
	maxConnAttempts int
	maxConnTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool // отвечает за управление пулом соединений
}

func NewPostgres(url string, opts ...option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:     _defaultMaxPoolSize,
		maxConnAttempts: _defaultMaxConnAttempts,
		maxConnTimeout:  _defaultMaxConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// разбор строки подключения (URL) к PostgreSQL и создание конфигурации пула соединений
	//postgres://user:password@host:port/db?param=value
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres NewPostgres pgxpool.ParseConfig: %w", err)
	}
	poolConfig.MaxConns = int32(pg.maxPoolSize)
	for range pg.maxConnAttempts {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}
		//l.Info("Postgres is trying to connect", slog.Int("attemts left:", pg.maxConnAttempts))
		pg.maxConnAttempts--
	}
	if err != nil {
		return nil, fmt.Errorf("postgres NewPostgres maxConnAttempts=0: %w", err)
	}
	return pg, nil
}

func (pg *Postgres) Close() {
	if pg.Pool != nil {
		pg.Pool.Close()
	}
}
