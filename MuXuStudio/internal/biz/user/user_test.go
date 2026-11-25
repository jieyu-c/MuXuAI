package user

import (
	"context"
	"errors"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
)

type registerRepo struct {
	existing  *User
	createErr error
	created   *User
}

func (r *registerRepo) FindByPhone(context.Context, Phone) (*User, error) {
	return r.existing, nil
}

func (r *registerRepo) Create(_ context.Context, u *User) error {
	r.created = u
	return r.createErr
}

func TestRegisterDuplicate(t *testing.T) {
	for _, concurrent := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing", true: "concurrent insert"}[concurrent], func(t *testing.T) {
			r := &registerRepo{}
			if concurrent {
				r.createErr = ErrUserAlreadyExists
			} else {
				r.existing = &User{ID: 1}
			}
			err := NewUserUsecase(r).Register(context.Background(), "13800138000", "test-password")
			if !errors.Is(err, ErrUserAlreadyExists) {
				t.Fatalf("expected duplicate business error, got %v", err)
			}
			public := kerrors.FromError(err)
			if public.Code != 400 || public.Reason != "REGISTER_FAILED" || public.Message != "无法完成注册，请检查填写信息；如已有账号，请登录或找回密码。" {
				t.Fatalf("unexpected public error: %v", public)
			}
			if !concurrent && r.created != nil {
				t.Fatal("existing account must not be written")
			}
		})
	}
}

func TestRegisterCreatesCanonicalPhoneAndHash(t *testing.T) {
	r := &registerRepo{}
	if err := NewUserUsecase(r).Register(context.Background(), "13800138000", "test-password"); err != nil {
		t.Fatal(err)
	}
	if r.created == nil || r.created.Phone.Value != "+8613800138000" || !r.created.PasswordHash.Matches("test-password") {
		t.Fatal("expected canonical phone and valid password hash")
	}
}
