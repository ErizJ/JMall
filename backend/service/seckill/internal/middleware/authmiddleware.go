package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			httpx.OkJson(w, map[string]string{"code": "401"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.secret), nil
		})
		if err != nil || !token.Valid {
			httpx.OkJson(w, map[string]string{"code": "401"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpx.OkJson(w, map[string]string{"code": "401"})
			return
		}

		userID, ok := claims["userId"]
		if !ok {
			httpx.OkJson(w, map[string]string{"code": "401"})
			return
		}

		ctx := context.WithValue(r.Context(), ctxutil.CtxKeyUserID, userID)
		next(w, r.WithContext(ctx))
	}
}
