package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func tmpDBInit(ctx context.Context, db *pgxpool.Pool, logger *zap.SugaredLogger) error {
	_, err := db.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS users (
				  id BIGSERIAL PRIMARY KEY,
				  login TEXT NOT NULL,
				  password TEXT NOT NULL
				);
				CREATE UNIQUE INDEX users_uniq_login ON users (login) NULLS NOT DISTINCT;
				CREATE TABLE IF NOT EXISTS orders (
				  id BIGSERIAL PRIMARY KEY,
				  login TEXT NOT NULL,
				  order_number TEXT NOT NULL,
				  withdrawn float8 DEFAULT 0,
				  uploaded_at TEXT NOT NULL,
				  CONSTRAINT fk_login
				    FOREIGN KEY(login)
				      REFERENCES users(login)
				        ON DELETE CASCADE
				);
				CREATE UNIQUE INDEX orders_uniq_order ON orders (order_number) NULLS NOT DISTINCT;
				CREATE TABLE IF NOT EXISTS balance (
				  id BIGSERIAL PRIMARY KEY,
				  login TEXT NOT NULL,
				  current float8,
				    CONSTRAINT fk_login
				      FOREIGN KEY(login)
				        REFERENCES users(login)
					  ON DELETE CASCADE
				);
				CREATE UNIQUE INDEX balance_uniq_login ON balance (login) NULLS NOT DISTINCT;
			`)
	if err != nil {
		logger.Errorln("Unable to init DB: ", err)
		return err
	}
	return nil
}
