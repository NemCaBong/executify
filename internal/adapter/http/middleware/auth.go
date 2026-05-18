package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/application/user"
	"github.com/NemCaBong/executify/internal/logger"
	"github.com/NemCaBong/executify/pkg/httperr"
)

const ContextKeyUserClaims = "user_claims"

// Auth validates the Bearer access token and injects parsed claims into the context.
func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			httperr.AbortUnauthorized(c, "missing or malformed authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &user.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			httperr.AbortUnauthorized(c, "invalid or expired access token")
			return
		}

		c.Set(ContextKeyUserClaims, claims)

		l := logger.FromContext(c.Request.Context()).With(zap.Int("user_id", claims.UserID))
		ctx := logger.WithContext(c.Request.Context(), l)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
