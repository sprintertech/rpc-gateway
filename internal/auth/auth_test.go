package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestURLTokenAuth(t *testing.T) {
	validToken := "valid_token"
	middleware := URLTokenAuth(validToken)

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
			name:           "Valid token",
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
