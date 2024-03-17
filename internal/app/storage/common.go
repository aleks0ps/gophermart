package storage

import "context"

type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

type Order struct{}
type Balance struct{}

type Register interface {
	Register(ctx context.Context, user *User) error
}

type Storage interface {
	Register
}
