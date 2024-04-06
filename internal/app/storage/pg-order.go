package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	myerror "github.com/aleks0ps/gophermart/internal/app/error"
)

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
		return myerror.ErrOrderLoaded
	}
	// Another user loaded this order
	return myerror.ErrOrderInUse
}

func (p *PGStorage) GetOrders(ctx context.Context, user *User) ([]*Order, error) {
	var count int
	err := p.DB.QueryRow(ctx, `SELECT count(order_number) FROM orders WHERE login=$1`, user.Login).Scan(&count)
	if err != nil {
		p.logger.Errorln(err.Error())
		return nil, err
	}
	if count == 0 {
		return nil, myerror.ErrNoOrders
	}
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
