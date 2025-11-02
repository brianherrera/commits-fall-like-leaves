package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianherrera/commits-fall-like-leaves/internal/service/haiku"
	"github.com/gin-gonic/gin"
)

type MockHaikuService struct {
	ResponseToReturn haiku.HaikuCommitResponse
	ErrorToReturn    error
}

func (m *MockHaikuService) CreateHaiku(ctx context.Context, request haiku.HaikuCommitRequest) (haiku.HaikuCommitResponse, error) {
	return m.ResponseToReturn, m.ErrorToReturn
}

func TestPostHaiku(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		requestBody        any
		mockResponse       haiku.HaikuCommitResponse
		mockError          error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "Successful request",
			requestBody: haiku.HaikuCommitRequest{
				CommitMessage: "fix: resolved login issue",
				Mood:          haiku.MoodReflective,
			},
			mockResponse: haiku.HaikuCommitResponse{
				Haiku: "Code changes merged in\nBugs squashed with precision now\nUsers rejoice, yay",
			},
			mockError:          nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Invalid JSON request",
			requestBody:        "invalid json",
			mockResponse:       haiku.HaikuCommitResponse{},
			mockError:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      InvalidRequest,
		},
		{
			name: "Service returns bad haiku request error",
			requestBody: haiku.HaikuCommitRequest{
				CommitMessage: "test commit",
				Mood:          "invalid_mood",
			},
			mockResponse:       haiku.HaikuCommitResponse{},
			mockError:          haiku.ErrBadHaikuRequest,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      InvalidRequest,
		},
		{
			name: "Service returns internal error",
			requestBody: haiku.HaikuCommitRequest{
				CommitMessage: "test commit",
				Mood:          haiku.MoodReflective,
			},
			mockResponse:       haiku.HaikuCommitResponse{},
			mockError:          errors.New("some internal error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      InternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockHaikuService{
				ResponseToReturn: tc.mockResponse,
				ErrorToReturn:    tc.mockError,
			}

			api := NewHaikuAPI(mockService)

			// Setup router
			router := gin.New()
			api.SetupRoutes(router)

			// Prepare request body
			var requestBody []byte
			var err error
			if str, ok := tc.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create request
			req, err := http.NewRequest("POST", "/haiku", bytes.NewBuffer(requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatusCode, w.Code)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tc.expectedStatusCode == http.StatusOK {
				// Check successful response
				if haiku, ok := response["haiku"].(string); !ok || haiku != tc.mockResponse.Haiku {
					t.Errorf("Expected haiku %q, got %q", tc.mockResponse.Haiku, haiku)
				}
			} else {
				// Check error response
				if errorMsg, ok := response["error"].(string); !ok || errorMsg != tc.expectedError {
					t.Errorf("Expected error %q, got %q", tc.expectedError, errorMsg)
				}
			}
		})
	}
}
