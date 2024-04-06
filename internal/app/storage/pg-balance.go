package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	myerror "github.com/aleks0ps/gophermart/internal/app/error"
	"github.com/jackc/pgx/v5"
)

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
		p.logger.Errorln(myerror.ErrInsufficientBalance.Error())
		return myerror.ErrInsufficientBalance
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
			return nil, myerror.ErrNoWithdrawals
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
