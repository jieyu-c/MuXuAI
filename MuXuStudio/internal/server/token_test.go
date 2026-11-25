package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/types/known/durationpb"
	"jieyuc.cn/muxu/studio/internal/conf"
	"jieyuc.cn/muxu/studio/internal/identity"
	"jieyuc.cn/muxu/studio/internal/service"
)

func tokenConfig() *conf.Server {
	return &conf.Server{
		Http: &conf.Server_HTTP{},
		Token: &conf.Server_Token{
			Key: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
			Ttl: durationpb.New(time.Hour),
		},
	}
}

func TestNewTokenManager(t *testing.T) {
	for _, name := range []string{"missing server", "missing token", "missing ttl", "bad base64", "wrong key size", "zero ttl", "invalid duration", "valid"} {
		t.Run(name, func(t *testing.T) {
			c := tokenConfig()
			switch name {
			case "missing server":
				c = nil
			case "missing token":
				c.Token = nil
			case "missing ttl":
				c.Token.Ttl = nil
			case "bad base64":
				c.Token.Key = "%%%"
			case "wrong key size":
				c.Token.Key = base64.StdEncoding.EncodeToString([]byte("short"))
			case "zero ttl":
				c.Token.Ttl = durationpb.New(0)
			case "invalid duration":
				c.Token.Ttl = &durationpb.Duration{Seconds: 1, Nanos: -1}
			}
			_, err := NewTokenManager(c)
			if (err == nil) != (name == "valid") {
				t.Fatalf("unexpected err=%v", err)
			}
		})
	}
}

func TestHTTPTokenAuthentication(t *testing.T) {
	c := tokenConfig()
	manager, err := NewTokenManager(c)
	if err != nil {
		t.Fatal(err)
	}
	valid, _, err := manager.Issue(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := base64.RawURLEncoding.DecodeString(valid)
	raw[len(raw)-1] ^= 1
	tampered := base64.RawURLEncoding.EncodeToString(raw)
	srv := NewHTTPServer(c, service.NewPingService(), &service.UserService{}, manager)
	// Exercise the same middleware chain with a handler that reads the identity.
	srv.Route("/").GET("/test/identity", func(ctx khttp.Context) error {
		handler := ctx.Middleware(func(ctx context.Context, _ any) (any, error) {
			id, ok := identity.UserID(ctx)
			if !ok {
				t.Error("missing authenticated identity")
			}
			return id, nil
		})
		result, err := handler(ctx, nil)
		if err != nil {
			return err
		}
		return ctx.Result(200, result)
	})
	for _, tc := range []struct {
		name, path, token string
		status            int
	}{
		{"missing", "/test/identity", "", 401},
		{"invalid", "/test/identity", "invalid", 401},
		{"tampered", "/test/identity", tampered, 401},
		{"valid", "/test/identity", valid, 200},
		{"existing protected route", "/api/v1/ping", valid, 200},
		{"existing protected route missing token", "/api/v1/ping", "", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				request.Header.Set("Authorization", "Bearer "+tc.token)
			}
			recorder := httptest.NewRecorder()
			srv.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body)
			}
			if tc.name == "valid" {
				var response struct {
					Data int `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Data != 42 {
					t.Fatalf("identity response=%s err=%v", recorder.Body, err)
				}
			}
		})
	}
	// Empty registration reaches required-field validation without a token.
	request := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	srv.ServeHTTP(recorder, request)
	if recorder.Code != 400 {
		t.Fatalf("public registration status=%d body=%s", recorder.Code, recorder.Body)
	}
}
