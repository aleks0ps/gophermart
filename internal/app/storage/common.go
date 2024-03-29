package storage

import "context"

type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

type Auth interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, user *User) error
}

type Order struct {
	Order      string `json:"order,omitempty" db:"order_number"`
	Number     string `json:"number,omitempty"` // alias
	Status     string `json:"status"`
	Accrual    string `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at" db:"uploaded_at"`
}

type Balance struct {
	Current   string `json:"current"`
	Withdrawn string `json:"withdrawn"`
}

type Buyer interface {
	LoadOrder(ctx context.Context, user *User, order *Order) error
	UpdateBalance(ctx context.Context, user *User, order *Order) error
	GetOrders(ctx context.Context, user *User) ([]*Order, error)
	GetBalance(ctx context.Context, user *User) (*Balance, error)
}

type Storage interface {
	Auth
	Buyer
}
