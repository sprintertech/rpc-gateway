package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestURLTokenAuth(t *testing.T) {
	validToken := "valid_token"
	tokenInfo := TokenInfo{
		Name:               "Test User",
		NumOfRequestPerSec: 1, // Changed from 60 per minute to 1 per second
	}
	tokenMap := map[string]TokenInfo{validToken: tokenInfo}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			url:            "/some/path/valid_token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid token with long path",
			url:            "/some/really/long/path/valid_token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			url:            "/some/path/invalid_token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing token",
			url:            "/some/path/",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := URLTokenAuth(tokenMap)
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v; got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestURLTokenAuthRateLimit(t *testing.T) {
	validToken := "valid_token"
	tokenInfo := TokenInfo{
		Name:               "Test User",
		NumOfRequestPerSec: 5, // Changed from 60 per minute to 1 per second
	}
	tokenMap := map[string]TokenInfo{validToken: tokenInfo}
	middleware := URLTokenAuth(tokenMap)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	url := "/some/path/valid_token"

	// Make requests up to the limit
	for i := 0; i < tokenInfo.NumOfRequestPerSec; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected OK for request %d, got %d", i, rr.Code)
		}
	}

	// This request should exceed the rate limit
	req, _ := http.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %v for rate limit exceeded; got %v", http.StatusTooManyRequests, rr.Code)
	}

	// Wait for a second to allow the rate limiter to reset
	time.Sleep(time.Second)

	// This request should now succeed
	req, _ = http.NewRequest("GET", url, nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %v after rate limit reset; got %v", http.StatusOK, rr.Code)
	}
}
