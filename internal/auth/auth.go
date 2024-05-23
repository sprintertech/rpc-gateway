package auth

import (
	"net/http"
	"strings"
)

func URLTokenAuth(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) < 2 || pathParts[len(pathParts)-1] != token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// Remove the token part from the path to forward the request to the next handler
			r.URL.Path = strings.Join(pathParts[:len(pathParts)-1], "/")
			next.ServeHTTP(w, r)
		})
	}
}
