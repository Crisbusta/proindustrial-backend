package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/gin-gonic/gin"
)

type ctxKey struct{}

// Init sets up the global slog handler. Call once at startup.
func Init(env string) {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))
}

// FromGin returns a logger enriched with request_id from a Gin context.
func FromGin(c *gin.Context) *slog.Logger {
	attrs := []any{}
	if id, ok := c.Get(middleware.RequestIDKey); ok {
		attrs = append(attrs, "request_id", id)
	}
	if uid := c.GetString(middleware.UserIDKey); uid != "" {
		attrs = append(attrs, "user_id", uid)
	}
	if cid := c.GetString(middleware.CompanyIDKey); cid != "" {
		attrs = append(attrs, "company_id", cid)
	}
	return slog.Default().With(attrs...)
}

// FromCtx returns a logger enriched with any values stored in a plain context.
func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
