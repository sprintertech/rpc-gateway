package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

type TokenInfo struct {
	Name               string `json:"name"`
	NumOfRequestPerSec int    `json:"numOfRequestPerSec"`
}

// Define a custom type for the context key
type ContextKeyType string

const TokenInfoKey ContextKeyType = "tokeninfo"

func URLTokenAuth(tokenToName map[string]TokenInfo) func(next http.Handler) http.Handler {
	limiters := make(map[string]*rate.Limiter)
	for token, info := range tokenToName {
		limiters[token] = rate.NewLimiter(rate.Limit(info.NumOfRequestPerSec), info.NumOfRequestPerSec)
		fmt.Printf("Configured limiter for %s, allowed %d requests per second\n",
			info.Name, info.NumOfRequestPerSec,
		)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) < 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token := pathParts[len(pathParts)-1]
			tInfo, validToken := tokenToName[token]
			if !validToken {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			limiter, exists := limiters[token]
			if !exists {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !limiter.Allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			// Remove the token part from the path
			r.URL.Path = strings.Join(pathParts[:len(pathParts)-1], "/")

			// Add the user's name to the request context
			ctx := context.WithValue(r.Context(), TokenInfoKey, tInfo)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
