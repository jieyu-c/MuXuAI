package user

import (
	"context"
	"errors"
	"jieyuc.cn/muxu/studio/internal/token"
	"log/slog"
	"testing"
	"time"

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
			err := newTestUsecase(t, r).Register(context.Background(), "13800138000", "test-password")
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
	if err := newTestUsecase(t, r).Register(context.Background(), "13800138000", "test-password"); err != nil {
		t.Fatal(err)
	}
	if r.created == nil || r.created.Phone.Value != "+8613800138000" || !r.created.PasswordHash.Matches("test-password") {
		t.Fatal("expected canonical phone and valid password hash")
	}
}

func newTestUsecase(t *testing.T, repo Repo) *Usecase {
	t.Helper()
	manager, err := token.New([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return NewUserUsecase(repo, manager, slog.Default())
}

func TestLoginIssuesVerifiableToken(t *testing.T) {
	password := Password{Value: "test-password"}
	if err := password.Hash(); err != nil {
		t.Fatal(err)
	}
	repo := &registerRepo{existing: &User{ID: 42, PasswordHash: password}}
	usecase := newTestUsecase(t, repo)
	ctx := context.Background()
	result, err := usecase.Login(ctx, "13800138000", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || !result.ExpireAt.After(time.Now()) {
		t.Fatal("missing login token or expiration")
	}
	id, err := usecase.tokenManager.Verify(ctx, result.Token)
	if err != nil || id != "42" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	if result, err := usecase.Login(ctx, "13800138000", "wrong-password"); err == nil || result != nil {
		t.Fatal("incorrect password must not issue a token")
	}
}
