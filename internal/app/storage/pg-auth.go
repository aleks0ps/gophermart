package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	myerror "github.com/aleks0ps/gophermart/internal/app/error"
	"github.com/aleks0ps/gophermart/internal/app/util"
)

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
				return myerror.ErrLoginAlreadyTaken
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
