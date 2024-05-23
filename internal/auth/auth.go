package auth

import (
	"net/http"
)

func URLTokenAuth(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authToken := r.URL.Query().Get("auth_token")
			if authToken == "" || authToken != token {
				w.WriteHeader(http.StatusUnauthorized)

				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
