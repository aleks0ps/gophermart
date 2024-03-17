package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/aleks0ps/gophermart/internal/app/gerror"
	"github.com/aleks0ps/gophermart/internal/app/util"
)

type PGStorage struct {
	DB     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func tmpDBInit(ctx context.Context, db *pgxpool.Pool, logger *zap.SugaredLogger) error {
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
				id BIGSERIAL PRIMARY KEY,
				login TEXT NOT NULL,
				password TEXT NOT NULL);
				CREATE UNIQUE INDEX unique_users ON users (login) NULLS NOT DISTINCT
				`)
	if err != nil {
		logger.Errorln("Unable to init DB: ", err)
		return err
	}
	return nil
}

func NewPGStorage(ctx context.Context, databaseDSN string, logger *zap.SugaredLogger) (*PGStorage, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseDSN)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	// Create tables
	if err := tmpDBInit(ctx, db, logger); err != nil {
		logger.Errorln(err)
	}
	return &PGStorage{DB: db, logger: logger}, nil
}

func (p *PGStorage) Register(ctx context.Context, user *User) error {
	hPassword, err := util.Hash(user.Password)
	if err != nil {
		return err
	}
	if _, err := p.DB.Exec(ctx, `INSERT INTO users(login, password) values ($1, $2)`, user.Login, hPassword); err != nil {
		p.logger.Errorln(err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Record already exists
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				// Wrap error
				return fmt.Errorf("%w", gerror.LoginAlreadyTaken)
			}
		}
		return err
	}
	return nil
}
