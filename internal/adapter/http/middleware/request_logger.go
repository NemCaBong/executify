package middleware

import (
	"time"

	"github.com/NemCaBong/executify/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const ContextKeyRequestID = "request_id"

// RequestLogger generates a unique request_id per request, seeds the context logger
// with it, and logs request start/completion with latency and HTTP status.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(ContextKeyRequestID, requestID)

		l := logger.FromContext(c.Request.Context()).With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		ctx := logger.WithContext(c.Request.Context(), l)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		l.Info("request started")

		c.Next()

		l.Info("request completed",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
