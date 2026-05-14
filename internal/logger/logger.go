package logger

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey struct{}

var global *zap.Logger

// Init builds a production zap logger (JSON, stdout) and sets it as the global.
// Log level is controlled by LOG_LEVEL env var (debug|info|warn|error; default: info).
func Init() *zap.Logger {
	level := zapcore.InfoLevel
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "ts"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		level,
	)
	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	global = l
	zap.ReplaceGlobals(l)
	return l
}

// WithContext stores l in ctx so downstream handlers and workers can retrieve it.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext returns the logger stored in ctx, falling back to the global logger.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	if global != nil {
		return global
	}
	return zap.L()
}
