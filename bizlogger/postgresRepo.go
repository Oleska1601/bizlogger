package bizlogger

import (
	"bizlogger/bizlogger/postgres"
	"context"
	"fmt"
	"os"
	"strings"
)

type PostgresRepo struct {
	db *postgres.Postgres
}

func newPostgresRepo(pg *postgres.Postgres) *PostgresRepo {
	return &PostgresRepo{db: pg}
}

func (pgRepo *PostgresRepo) createTables(filepath string) error {
	queriesBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	queries := strings.Split(string(queriesBytes), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query) //убрать лишние пробелы
		if query == "" {
			continue
		}
		_, err := pgRepo.db.Pool.Exec(context.Background(), query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pgRepo *PostgresRepo) insertLogInDB(ctx context.Context, bizlog *BizLog) error {
	sql, args, err := pgRepo.db.Builder.
		Insert("business_logs").
		Columns("timestamp", "event_type", "entity", "username",
			"user_role", "context", "entity_id", "old_value", "new_value", "description").
		Values(bizlog.Timestamp, bizlog.EventType, bizlog.Entity, bizlog.Username,
			bizlog.UserRole, bizlog.Context, bizlog.EntityID, bizlog.OldValue, bizlog.NewValue, bizlog.Description).
		ToSql()
	if err != nil {
		return fmt.Errorf("insertLog pgRepo.db.Builder.Insert: %w", err)
	}
	_, err = pgRepo.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insertLog pgRepo.db.Pool.Exec: %w", err)
	}
	return nil
}
