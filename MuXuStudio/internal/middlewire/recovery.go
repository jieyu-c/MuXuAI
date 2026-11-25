package middlewire

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
)

// Recovery preserves Kratos panic recovery while omitting request bodies from
// its logs. Login and registration requests contain plaintext credentials.
func Recovery() middleware.Middleware {
	logger := slog.New(withoutRequest{handler: slog.Default().Handler()})
	return recovery.Recovery(recovery.WithLogger(logger))
}

type withoutRequest struct{ handler slog.Handler }

func (h withoutRequest) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h withoutRequest) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "request" {
			clean.AddAttrs(attr)
		}
		return true
	})
	return h.handler.Handle(ctx, clean)
}

func (h withoutRequest) WithAttrs(attrs []slog.Attr) slog.Handler {
	filtered := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Key != "request" {
			filtered = append(filtered, attr)
		}
	}
	return withoutRequest{handler: h.handler.WithAttrs(filtered)}
}

func (h withoutRequest) WithGroup(name string) slog.Handler {
	return withoutRequest{handler: h.handler.WithGroup(name)}
}
