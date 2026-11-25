package data

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	userpkg "jieyuc.cn/muxu/studio/internal/biz/user"
)

// User is the database model for a phone-based account.
type User struct {
	gorm.Model
	// Phone stores the canonical E.164 number, including the leading plus sign.
	Phone string `gorm:"size:16;not null;uniqueIndex:idx_users_phone"`
	// PasswordHash stores the encoded password hash.
	PasswordHash string `gorm:"type:text;not null" json:"-"`
}

func toDomain(user User) *userpkg.User {
	var deletedAt *time.Time
	if user.DeletedAt.Valid {
		deletedAt = &user.DeletedAt.Time
	} else {
		deletedAt = nil
	}
	return &userpkg.User{
		ID:           user.ID,
		Phone:        userpkg.Phone{Value: user.Phone},
		PasswordHash: userpkg.Password{Value: user.PasswordHash},
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func toPO(bizUser *userpkg.User) *User {
	var deletedAt gorm.DeletedAt
	if bizUser.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *bizUser.DeletedAt, Valid: true}
	}
	return &User{
		Model: gorm.Model{
			ID: bizUser.ID, CreatedAt: bizUser.CreatedAt,
			UpdatedAt: bizUser.UpdatedAt, DeletedAt: deletedAt,
		},
		Phone:        bizUser.Phone.Value,
		PasswordHash: bizUser.PasswordHash.Value,
	}
}

type UserRepo struct {
	data *Data
}

func NewUserRepo(data *Data) userpkg.Repo {
	return &UserRepo{
		data: data,
	}
}

var _ userpkg.Repo = (*UserRepo)(nil)

func (u *UserRepo) FindByPhone(ctx context.Context, phone userpkg.Phone) (*userpkg.User, error) {
	user, err := gorm.G[User](u.data.db).Where("phone = ?", phone.Value).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return toDomain(user), nil
}

func (u *UserRepo) Create(ctx context.Context, user *userpkg.User) error {
	err := gorm.G[User](u.data.db).Create(ctx, toPO(user))
	return err
}
