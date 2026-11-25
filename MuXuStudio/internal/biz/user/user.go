package user

import (
	"context"
	"time"
)

type User struct {
	ID           uint
	Phone        string
	PasswordHash Password
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type UserRepo interface {
	FindByPhone(context.Context, string) (*User, error)
	Create(context.Context, *User) error
}

type UserUsecase struct {
}
