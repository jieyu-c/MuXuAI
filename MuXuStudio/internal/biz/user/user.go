package user

import (
	"context"
	"time"
)

type User struct {
	ID           uint
	Phone        Phone
	PasswordHash Password
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type Repo interface {
	FindByPhone(context.Context, Phone) (*User, error)
	Create(context.Context, *User) error
}

type Usecase struct {
	repo Repo
}

func NewUserUsecase(repo Repo) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Register(ctx context.Context, phone, password string) error {
	phoneVO := Phone{phone}
	err := phoneVO.Standardize()
	if err != nil {
		return err
	}

	user, err := u.repo.FindByPhone(ctx, phoneVO)

	if err != nil {
		return err
	}
	if user != nil {
		return ErrUserAlreadyExists
	}

	passwordVO := Password{password}
	err = passwordVO.Hash()
	if err != nil {
		return err
	}

	registerUser := &User{
		Phone:        phoneVO,
		PasswordHash: passwordVO,
	}
	err = u.repo.Create(ctx, registerUser)
	if err != nil {
		return err
	}
	return nil
}
