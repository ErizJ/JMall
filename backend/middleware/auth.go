package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AuthMiddleware validates the user session / JWT token.
// Each service's generated authmiddleware.go should delegate here.
type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement JWT validation
		// 1. Read Authorization header (Bearer <token>)
		// 2. Parse and validate token using config.Auth.Secret
		// 3. Inject user_id into context
		// 4. Return 401 JSON on failure

		// Placeholder: allow all requests through
		_ = httpx.OkJson // ensure httpx is imported
		next(w, r)
	}
}
