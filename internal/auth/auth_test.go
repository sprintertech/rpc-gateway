package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUrlTokenAuth(t *testing.T) {
	validToken := "valid_token"
	middleware := UrlTokenAuth(validToken)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			url:            "/?auth_token=valid_token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			url:            "/?auth_token=invalid_token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing token",
			url:            "/",
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
