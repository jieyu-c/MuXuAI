package middlewire

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	v1 "jieyuc.cn/muxu/studio/api/user/v1"
)

func TestRecoveryOmitsCredentials(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	defer slog.SetDefault(previous)
	handler := Recovery()(func(context.Context, any) (any, error) { panic("test panic") })
	_, err := handler(context.Background(), &v1.RegisterRequest{Phone: "13800138000", Password: "secret-test-password"})
	if kerrors.Code(err) != 500 {
		t.Fatalf("expected 500, got %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if _, exists := record["request"]; exists {
		t.Fatal("request logged")
	}
	if bytes.Contains(output.Bytes(), []byte("secret-test-password")) || bytes.Contains(output.Bytes(), []byte("13800138000")) {
		t.Fatal("credentials logged")
	}
	if record["panic"] != "test panic" || record["stack"] == nil {
		t.Fatal("missing panic diagnostics")
	}
}
