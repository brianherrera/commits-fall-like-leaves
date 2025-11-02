package haiku

import (
	"context"
	"errors"
	"testing"

	"github.com/brianherrera/commits-fall-like-leaves/internal/clients/bedrock"
)

// MockBedrockClient implements the BedrockClient interface for testing
type MockBedrockClient struct {
	ResponseToReturn string
	ErrorToReturn    error
}

func (m *MockBedrockClient) InvokeClaude(ctx context.Context, prompt string, opts *bedrock.ClaudeOptions) (string, error) {
	return m.ResponseToReturn, m.ErrorToReturn
}

func TestCreateHaiku(t *testing.T) {
	mockError := errors.New("bedrock API error")

	tests := []struct {
		name           string
		commitMessage  string
		mood           Mood
		expectedPrompt string
		mockResponse   string
		mockError      error
		expectError    bool
		errorIs        error
	}{
		{
			name:           "Default mood",
			commitMessage:  "fix: resolved login issue",
			mood:           Mood(""), // Empty to test default
			expectedPrompt: "Create a reflective haiku from this commit message: fix: resolved login issue",
			mockResponse:   "Code changes merged\nBugs squashed with precision now\nUsers rejoice, yay",
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "Custom mood",
			commitMessage:  "refactor: optimize database queries",
			mood:           MoodTechnical,
			expectedPrompt: "Create a technical haiku from this commit message: refactor: optimize database queries",
			mockResponse:   "Technical changes\nRefactoring the codebase\nPerformance improved",
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "Invalid mood",
			commitMessage:  "feat: add easter egg",
			mood:           Mood("silly"), // Not a valid mood
			expectedPrompt: "",
			mockResponse:   "",
			mockError:      nil,
			expectError:    true,
			errorIs:        ErrBadHaikuRequest,
		},
		{
			name:           "Bedrock error",
			commitMessage:  "fix: resolved login issue",
			mood:           MoodReflective,
			expectedPrompt: "Create a reflective haiku from this commit message: fix: resolved login issue",
			mockResponse:   "",
			mockError:      mockError,
			expectError:    true,
			errorIs:        ErrCreateHaiku,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockBedrockClient{
				ResponseToReturn: tc.mockResponse,
				ErrorToReturn:    tc.mockError,
			}

			service := NewHaikuService(mockClient)
			request := HaikuCommitRequest{
				CommitMessage: tc.commitMessage,
				Mood:          tc.mood,
			}

			response, err := service.CreateHaiku(context.Background(), request)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else {
					// Check error details if expecting an error
					if tc.errorIs != nil && !errors.Is(err, tc.errorIs) {
						t.Errorf("Expected error to wrap %v, got %v", tc.errorIs, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				} else {
					// Check response if not expecting an error
					if response.Haiku != tc.mockResponse {
						t.Errorf("Expected haiku %q, got %q", tc.mockResponse, response.Haiku)
					}
				}
			}
		})
	}
}
