package storage

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PGStorage struct {
	DB     *pgxpool.Pool
	logger *zap.SugaredLogger
}

type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

type Auth interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, user *User) error
}

type Order struct {
	Order      string      `json:"order,omitempty" db:"order_number"`
	Number     string      `json:"number,omitempty"` // alias
	Withdrawn  json.Number `json:"sum"`
	Status     string      `json:"status,omitempty"`
	Accrual    json.Number `json:"accrual,omitempty"`
	UploadedAt string      `json:"uploaded_at,omitempty" db:"uploaded_at"`
}

type Balance struct {
	Current   json.Number `json:"current"`
	Withdrawn json.Number `json:"withdrawn"`
}

type Buyer interface {
	LoadOrder(ctx context.Context, user *User, order *Order) error
	CheckWithdrawn(ctx context.Context, user *User, order *Order) error
	BalanceInit(ctx context.Context, user *User) error
	BalanceIncrease(ctx context.Context, user *User, order *Order) error
	BalanceDecrease(ctx context.Context, user *User, order *Order) error
	GetOrders(ctx context.Context, user *User) ([]*Order, error)
	GetWithdrawals(ctx context.Context, user *User) ([]*Order, error)
	GetBalance(ctx context.Context, user *User) (*Balance, error)
}

type Storage interface {
	Auth
	Buyer
}

func jsonNumberToFloat64(number json.Number) (float64, error) {
	if number.String() == "" {
		return float64(0), nil
	}
	f64, err := number.Float64()
	if err != nil {
		return float64(0), err
	}
	return f64, nil
}

func jsonNumberToFloat32(number json.Number) (float32, error) {
	f64, err := jsonNumberToFloat64(number)
	return float32(f64), err
}

func isFloat(number json.Number) error {
	_, err := number.Float64()
	if err != nil {
		return err
	}
	return nil
}

func isFloatS(str string) error {
	_, err := strconv.ParseFloat(str, 64)
	return err
}
