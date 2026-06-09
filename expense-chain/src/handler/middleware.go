package handler

import (
	"context"
	"net/http"
	"strings"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type ctxKey string

const claimsKey ctxKey = "claims"

// AuthMiddleware validates JWTs and enforces role-based access.
type AuthMiddleware struct {
	auth *service.AuthService
}

func NewAuthMiddleware(auth *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

// Protect wraps a handler: requires a valid access token and (optionally) one of the given roles.
// Empty roles list = any authenticated user. ADMIN always passes.
func (m *AuthMiddleware) Protect(next http.HandlerFunc, roles ...model.Role) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := m.auth.ParseToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if claims.TokenType != "access" {
			writeError(w, http.StatusUnauthorized, "not an access token")
			return
		}

		if !roleAllowed(claims.Role, roles) {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// roleAllowed returns true if role is ADMIN, the allow-list is empty, or role is listed.
func roleAllowed(role model.Role, allowed []model.Role) bool {
	if role == model.RoleAdmin {
		return true
	}
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if role == a {
			return true
		}
	}
	return false
}

// ClaimsFrom extracts the authenticated user's claims from the request context.
func ClaimsFrom(r *http.Request) (*service.Claims, bool) {
	claims, ok := r.Context().Value(claimsKey).(*service.Claims)
	return claims, ok
}
