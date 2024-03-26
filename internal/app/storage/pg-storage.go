package storage

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	myerror "github.com/aleks0ps/gophermart/internal/app/error"
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// DuplicateTable
			if pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code) {
				return &PGStorage{DB: db, logger: logger}, nil
			}
			return nil, err
		}
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
				return myerror.LoginAlreadyTaken
			}
		}
		return err
	}
	return nil
}

func (p *PGStorage) Login(ctx context.Context, user *User) error {
	var hPassword string
	err := p.DB.QueryRow(ctx, `SELECT password FROM users WHERE login=$1`, user.Login).Scan(&hPassword)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	err = util.CheckPasswordHash(hPassword, user.Password)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	return nil
}

func (p *PGStorage) LoadOrder(ctx context.Context, user *User, order *Order) error {
	var login, orderNumber string
	// Check if order alredy exists
	err := p.DB.QueryRow(ctx, `SELECT login, order_number FROM orders WHERE order_number=$1`, order.Order).Scan(&login, &orderNumber)
	if err != nil {
		// Insert data if there is no entries
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := p.DB.Exec(ctx, `INSERT INTO orders(login,order_number,uploaded_at) values ($1,$2,$3)`, user.Login, order.Order, order.UploadedAt)
			if err != nil {
				p.logger.Errorln(err.Error())
				return err
			}
			return nil
		} else {
			p.logger.Errorln(err.Error())
			return err
		}
	}
	// User alredy loaded it
	if login == user.Login {
		return myerror.OrderLoaded
	}
	// Another user loaded this order
	return myerror.OrderInUse
}

func (p *PGStorage) UpdateBalance(ctx context.Context, user *User, order *Order) error {
	current, err := strconv.ParseFloat(order.Accrual, 64)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	var login string
	err = p.DB.QueryRow(ctx, `SELECT login FROM balance WHERE login=$1`, user.Login).Scan(&login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := p.DB.Exec(ctx, `INSERT INTO balance (login, current) values ($1,$2)`, user.Login, current)
			if err != nil {
				p.logger.Errorln(err.Error())
				return err
			}
		} else {
			p.logger.Errorln(err.Error())
			return err
		}
	}
	_, err = p.DB.Exec(ctx, `UPDATE balance SET current = current + $1 WHERE login = $2`, current, user.Login)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	return nil
}

func (p *PGStorage) GetOrders(ctx context.Context, user *User) ([]*Order, error) {
	rows, err := p.DB.Query(ctx, `SELECT order_number, uploaded_at FROM orders WHERE login=$1`, user.Login)
	if err != nil {
		p.logger.Errorln(err.Error())
		return nil, err
	}
	defer rows.Close()
	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.Order, &order.UploadedAt)
		if err != nil {
			p.logger.Errorln(err.Error())
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		p.logger.Errorln(err.Error())
		return orders, err
	}
	return orders, nil
}
