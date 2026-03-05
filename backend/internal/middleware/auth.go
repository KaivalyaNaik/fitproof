package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/KaivalyaNaik/fitproof/pkg/respond"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

func Authenticate(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "missing access token")
				return
			}

			var claims jwt.RegisteredClaims
			token, err := jwt.ParseWithClaims(cookie.Value, &claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			}, jwt.WithExpirationRequired())
			if err != nil || !token.Valid {
				respond.Error(w, http.StatusUnauthorized, "invalid or expired access token")
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid token subject")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
