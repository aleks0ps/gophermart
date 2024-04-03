package storage

import (
	"context"
	"encoding/json"
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
			withdrawn, err := jsonNumberToFloat64(order.Withdrawn)
			if err != nil {
				p.logger.Errorln(err.Error())
				return err
			}
			_, err = p.DB.Exec(ctx, `INSERT INTO orders(login,order_number,withdrawn,uploaded_at) values ($1,$2,$3,$4)`,
				user.Login, order.Order, withdrawn, order.UploadedAt)
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

func (p *PGStorage) CheckWithdrawn(ctx context.Context, user *User, order *Order) error {
	var sCurrent string
	var withdrawn float64
	if order.Withdrawn.String() != "" {
		f, err := order.Withdrawn.Float64()
		if err != nil {
			p.logger.Errorln(err.Error())
			return err
		}
		withdrawn = f
	}
	// find current balance
	err := p.DB.QueryRow(ctx, `SELECT current FROM balance WHERE login=$1`, user.Login).Scan(&sCurrent)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	current, err := strconv.ParseFloat(sCurrent, 64)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	// check that we have enough balance to withdraw
	if current < withdrawn {
		p.logger.Errorln(myerror.InsufficientBalance.Error())
		return myerror.InsufficientBalance
	}
	return nil
}

func (p *PGStorage) BalanceInit(ctx context.Context, user *User) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO balance (login, current) VALUES ($1,$2)`, user.Login, 0)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	return nil
}

func (p *PGStorage) BalanceIncrease(ctx context.Context, user *User, order *Order) error {
	accrualPoints, err := jsonNumberToFloat64(order.Accrual)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	_, err = p.DB.Exec(ctx, `UPDATE balance SET current = balance.current+$1 WHERE login=$2`, accrualPoints, user.Login)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	return nil
}

func (p *PGStorage) BalanceDecrease(ctx context.Context, user *User, order *Order) error {
	if err := p.CheckWithdrawn(ctx, user, order); err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	withdrawn, err := jsonNumberToFloat64(order.Withdrawn)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	_, err = p.DB.Exec(ctx, `UPDATE balance SET current = balance.current-$1 WHERE login=$2`, withdrawn, user.Login)
	if err != nil {
		p.logger.Errorln(err.Error())
		return err
	}
	return nil
}

func (p *PGStorage) GetOrders(ctx context.Context, user *User) ([]*Order, error) {
	rows, err := p.DB.Query(ctx, `SELECT order_number, uploaded_at FROM orders WHERE login=$1 ORDER BY uploaded_at`, user.Login)
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

func (p *PGStorage) GetBalance(ctx context.Context, user *User) (*Balance, error) {
	var current, withdrawn string
	var count int
	var balance Balance
	// by default balance is zero
	err := p.DB.QueryRow(ctx, `SELECT current FROM balance WHERE login=$1`, user.Login).Scan(&current)
	if err != nil {
		//	if !errors.Is(err, pgx.ErrNoRows) {
		p.logger.Errorln(err.Error())
		return nil, err
		//	}
	}
	if err := isFloatS(current); err != nil {
		p.logger.Errorln(err.Error())
		return nil, err
	}
	// save balance
	balance.Current = json.Number(current)
	err = p.DB.QueryRow(ctx, `SELECT count(withdrawn) FROM orders WHERE login=$1`, user.Login).Scan(&count)
	if err != nil {
		p.logger.Errorln(err.Error())
		return nil, err
	}
	if count > 0 {
		// sum all withdrawals
		err = p.DB.QueryRow(ctx, `SELECT sum(withdrawn) FROM orders WHERE login=$1`, user.Login).Scan(&withdrawn)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				p.logger.Errorln(err.Error())
				return nil, err
			}
		}
		if err := isFloatS(withdrawn); err != nil {
			p.logger.Errorln(err.Error())
			return nil, err
		}
		balance.Withdrawn = json.Number(withdrawn)
	} else {
		balance.Withdrawn = json.Number("0")
	}
	return &balance, nil
}

func (p *PGStorage) GetWithdrawals(ctx context.Context, user *User) ([]*Order, error) {
	rows, err := p.DB.Query(ctx, `SELECT order_number,withdrawn,uploaded_at 
				      FROM orders 
				      WHERE login=$1 AND withdrawn > 0`, user.Login)
	if err != nil {
		p.logger.Errorln(err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerror.NoWithdrawals
		}
		return nil, err
	}
	defer rows.Close()
	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.Order, &order.Withdrawn, &order.UploadedAt)
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
