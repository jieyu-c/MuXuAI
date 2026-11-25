package data

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/schema"
	userpkg "jieyuc.cn/muxu/studio/internal/biz/user"
)

func TestMapUserCreateError(t *testing.T) {
	phoneConflict := &pgconn.PgError{Code: "23505", ConstraintName: "idx_users_phone", Detail: "sensitive phone"}
	for _, tc := range []struct {
		name      string
		err       error
		duplicate bool
	}{
		{"success", nil, false},
		{"phone", phoneConflict, true},
		{"wrapped phone", fmt.Errorf("insert: %w", phoneConflict), true},
		{"primary key", &pgconn.PgError{Code: "23505", ConstraintName: "users_pkey"}, false},
		{"other constraint", &pgconn.PgError{Code: "23502", ConstraintName: "idx_users_phone"}, false},
		{"storage failure", errors.New("database unavailable"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := mapUserCreateError(tc.err)
			if tc.duplicate {
				if got != userpkg.ErrUserAlreadyExists || errors.Is(got, phoneConflict) {
					t.Fatalf("expected sanitized business error, got %v", got)
				}
			} else if got != tc.err {
				t.Fatalf("unrelated error changed: %v", got)
			}
		})
	}
}

func TestPhoneUniqueIndex(t *testing.T) {
	s, err := schema.Parse(&User{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range s.ParseIndexes() {
		if index.Name == "idx_users_phone" && index.Class == "UNIQUE" && len(index.Fields) == 1 && index.Fields[0].DBName == "phone" {
			return
		}
	}
	t.Fatal("phone conflict mapping requires the idx_users_phone unique index")
}
