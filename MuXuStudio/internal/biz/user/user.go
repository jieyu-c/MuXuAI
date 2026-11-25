package user

import (
	"context"
	"log/slog"
	"time"

	"jieyuc.cn/muxu/studio/internal/token"
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
	repo         Repo
	tokenManager *token.Manager
	logger       *slog.Logger
}

func NewUserUsecase(repo Repo, tokenManager *token.Manager, logger *slog.Logger) *Usecase {
	return &Usecase{repo: repo, tokenManager: tokenManager, logger: logger}
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
		u.logger.Warn("phone is already registered")
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

type LoginResult struct {
	Token    string
	ExpireAt time.Time
}

func (u *Usecase) Login(ctx context.Context, phone, password string) (*LoginResult, error) {
	phoneVO := Phone{phone}
	err := phoneVO.Standardize()
	if err != nil {
		return nil, err
	}
	user, err := u.repo.FindByPhone(ctx, phoneVO)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotExists
	}
	if !user.PasswordHash.Matches(password) {
		return nil, ErrPasswordWrong
	}
	u.logger.Info("user %s", user)

	authToken, expiredAt, err := u.tokenManager.Issue(ctx, int(user.ID))
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:    authToken,
		ExpireAt: expiredAt,
	}, nil
}
