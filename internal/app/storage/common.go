package storage

import "context"

type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

type Register interface {
	Register(ctx context.Context, user *User) error
}

type Login interface {
	Login(ctx context.Context, user *User) error
}

type Order struct{}
type Balance struct{}

type Storage interface {
	Register
	Login
}
