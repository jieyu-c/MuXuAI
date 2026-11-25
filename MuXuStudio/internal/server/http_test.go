package server

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
)

func TestResponseEncoding(t *testing.T) {
	for _, tc := range []struct {
		name         string
		value        any
		status       int
		code         int
		msg          string
		errorEncoder bool
	}{
		{name: "success", value: "result", status: 200, msg: "ok"},
		{name: "empty success", status: 200, msg: "ok"},
		{name: "response error", value: kerrors.BadRequest("INVALID", "invalid input"), status: 400, code: 400, msg: "invalid input"},
		{name: "handler error", value: kerrors.NotFound("MISSING", "missing"), status: 404, code: 404, msg: "missing", errorEncoder: true},
		{name: "unknown error", value: stderrors.New("failed"), status: 500, code: 500, msg: "failed", errorEncoder: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.errorEncoder {
				encodeErrorFunc(recorder, req, tc.value.(error))
			} else if err := encodeResponseFunc(recorder, req, tc.value); err != nil {
				t.Fatal(err)
			}
			var response HttpResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tc.status || response.Code != tc.code || response.Msg != tc.msg {
				t.Fatalf("status=%d response=%+v", recorder.Code, response)
			}
			if response.ServerTime.IsZero() {
				t.Fatal("missing server time")
			}
			if tc.code != 0 && response.Data != nil {
				t.Fatalf("error data=%v", response.Data)
			}
			if tc.code == 0 && response.Data != tc.value {
				t.Fatalf("data=%v want=%v", response.Data, tc.value)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type=%q", got)
			}
		})
	}
}
