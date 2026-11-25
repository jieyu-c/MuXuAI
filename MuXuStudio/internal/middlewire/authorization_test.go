package middlewire

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport"
	"jieyuc.cn/muxu/studio/internal/identity"
)

type testHeader struct{ http.Header }

func (h testHeader) Keys() []string {
	var keys []string
	for k := range h.Header {
		keys = append(keys, k)
	}
	return keys
}

type testTransport struct {
	transport.Transporter
	request *http.Request
}

func (t testTransport) Request() *http.Request          { return t.request }
func (t testTransport) PathTemplate() string            { return t.request.URL.Path }
func (t testTransport) RequestHeader() transport.Header { return testHeader{t.request.Header} }

type testVerifier func(context.Context, string) (string, error)

func (v testVerifier) Verify(ctx context.Context, token string) (string, error) { return v(ctx, token) }

func TestAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, header, userID string
		verifyErr                          error
		public, allowed                    bool
	}{
		{name: "public", method: "POST", path: "/login?source=web", public: true, allowed: true},
		{name: "different method", method: "GET", path: "/login", public: true},
		{name: "different path", method: "POST", path: "/login/extra", public: true},
		{name: "missing", method: "GET", path: "/private"},
		{name: "wrong scheme", method: "GET", path: "/private", header: "Basic abc"},
		{name: "missing token", method: "GET", path: "/private", header: "Bearer"},
		{name: "extra fields", method: "GET", path: "/private", header: "Bearer abc def"},
		{name: "invalid or expired", method: "GET", path: "/private", header: "Bearer abc", verifyErr: errors.New("expired")},
		{name: "empty identity", method: "GET", path: "/private", header: "Bearer abc"},
		{name: "nonnumeric identity", method: "GET", path: "/private", header: "Bearer abc", userID: "abc"},
		{name: "zero identity", method: "GET", path: "/private", header: "Bearer abc", userID: "0"},
		{name: "negative identity", method: "GET", path: "/private", header: "Bearer abc", userID: "-1"},
		{name: "overflow identity", method: "GET", path: "/private", header: "Bearer abc", userID: "99999999999999999999999999999"},
		{name: "valid", method: "GET", path: "/private", header: "bearer abc", userID: "42", allowed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.header != "" {
				request.Header.Set("Authorization", tc.header)
			}
			ctx := transport.NewServerContext(context.Background(), testTransport{request: request})
			checked, called := false, false
			verifier := testVerifier(func(_ context.Context, token string) (string, error) {
				checked = true
				if token != "abc" {
					t.Fatalf("token=%q", token)
				}
				return tc.userID, tc.verifyErr
			})
			handler := Authorization(verifier, Route{Method: "POST", Path: "/login"})(func(ctx context.Context, _ any) (any, error) {
				called = true
				if !tc.public {
					child, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()
					if id, ok := identity.UserID(child); !ok || strconv.Itoa(id) != tc.userID {
						t.Fatalf("identity=%d ok=%v", id, ok)
					}
				}
				return "ok", nil
			})
			_, err := handler(ctx, nil)
			if called != tc.allowed {
				t.Fatalf("called=%v err=%v", called, err)
			}
			if tc.allowed && err != nil {
				t.Fatal(err)
			}
			if !tc.allowed && kerrors.Code(err) != 401 {
				t.Fatalf("expected 401, got %v", err)
			}
			if tc.name == "public" && checked {
				t.Fatal("public route verified token")
			}
			if _, ok := identity.UserID(ctx); ok {
				t.Fatal("parent context was modified")
			}
		})
	}
}

func TestAuthorizationFailsClosed(t *testing.T) {
	handler := Authorization(nil)(func(context.Context, any) (any, error) { t.Fatal("unexpected handler call"); return nil, nil })
	if _, err := handler(context.Background(), nil); kerrors.Code(err) != 401 {
		t.Fatalf("expected 401, got %v", err)
	}
	request := httptest.NewRequest("GET", "/private", nil)
	request.Header.Set("Authorization", "Bearer abc")
	ctx := transport.NewServerContext(context.Background(), testTransport{request: request})
	if _, err := handler(ctx, nil); kerrors.Code(err) != 401 {
		t.Fatalf("expected 401 with nil verifier, got %v", err)
	}
	request.Header.Add("Authorization", "Bearer def")
	if _, err := handler(ctx, nil); kerrors.Code(err) != 401 {
		t.Fatalf("expected 401 for duplicate headers, got %v", err)
	}
}
