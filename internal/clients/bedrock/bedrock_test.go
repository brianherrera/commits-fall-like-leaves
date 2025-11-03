package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go"
)

type MockBedrockRuntime struct {
	InvokeModelFunc func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

func (m *MockBedrockRuntime) InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
	if m.InvokeModelFunc != nil {
		return m.InvokeModelFunc(ctx, params, optFns...)
	}
	return nil, errors.New("InvokeModelFunc not implemented")
}

func TestInvokeClaudeValidation(t *testing.T) {
	mock := &MockBedrockRuntime{
		InvokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
			response := ClaudeResponse{Content: []ContentBlock{
				{
					Text: "Test response",
					Type: "text",
				},
			}}
			responseBytes, _ := json.Marshal(response)
			return &bedrockruntime.InvokeModelOutput{
				Body: responseBytes,
			}, nil
		},
	}

	client := NewBedrockClient(mock)

	tests := []struct {
		name          string
		prompt        string
		opts          *ClaudeOptions
		expectError   bool
		errorContains string
	}{
		{
			name:          "Empty prompt",
			prompt:        "",
			opts:          nil,
			expectError:   true,
			errorContains: "prompt cannot be empty",
		},
		{
			name:        "Valid prompt with nil options",
			prompt:      "Hello, world!",
			opts:        nil,
			expectError: false,
		},
		{
			name:   "Valid prompt with valid options",
			prompt: "Hello, world!",
			opts: &ClaudeOptions{
				MaxTokens:   100,
				Temperature: 0.7,
			},
			expectError: false,
		},
		{
			name:   "Valid prompt with zero temperature",
			prompt: "Hello, world!",
			opts: &ClaudeOptions{
				MaxTokens:   100,
				Temperature: 0.0, // Should be valid
			},
			expectError: false,
		},
		{
			name:   "Valid prompt with invalid temperature",
			prompt: "Hello, world!",
			opts: &ClaudeOptions{
				MaxTokens:   100,
				Temperature: 1.5, // Invalid value - will be ignored
			},
			expectError: false, // Not an error, just uses default
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.InvokeClaude(context.Background(), tc.prompt, tc.opts)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectError && err != nil && tc.errorContains != "" {
				if !errors.Is(err, ErrInvalidRequest) {
					t.Errorf("Expected error to wrap ErrInvalidRequest")
				}
				if !errors.Is(err, ErrInvalidRequest) {
					t.Errorf("Error doesn't match expected type: %v", err)
				}
				if err.Error() == "" || !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message doesn't contain expected text %q: %v", tc.errorContains, err)
				}
			}
		})
	}
}

// TestInvokeClaudeErrors tests error handling
func TestInvokeClaudeErrors(t *testing.T) {
	tests := []struct {
		name           string
		mockError      error
		expectedError  error
		errorCodeCheck func(error) bool
	}{
		{
			name: "Validation error",
			mockError: &smithy.GenericAPIError{
				Code:    ValidationExceptionCode,
				Message: "Validation failed",
			},
			expectedError:  ErrValidation,
			errorCodeCheck: func(err error) bool { return errors.Is(err, ErrValidation) },
		},
		{
			name: "Resource not found error",
			mockError: &smithy.GenericAPIError{
				Code:    ResourceNotFoundExceptionCode,
				Message: "Model not found",
			},
			expectedError:  ErrModelUnavailable,
			errorCodeCheck: func(err error) bool { return errors.Is(err, ErrModelUnavailable) },
		},
		{
			name: "Throttling error",
			mockError: &smithy.GenericAPIError{
				Code:    ThrottlingExceptionCode,
				Message: "Request was throttled",
			},
			expectedError:  ErrThrottling,
			errorCodeCheck: func(err error) bool { return errors.Is(err, ErrThrottling) },
		},
		{
			name: "Quota exceeded error",
			mockError: &smithy.GenericAPIError{
				Code:    ServiceQuotaExceededExceptionCode,
				Message: "Service quota exceeded",
			},
			expectedError:  ErrQuotaExceeded,
			errorCodeCheck: func(err error) bool { return errors.Is(err, ErrQuotaExceeded) },
		},
		{
			name: "Unknown error",
			mockError: &smithy.GenericAPIError{
				Code:    "UnknownError",
				Message: "Unknown error occurred",
			},
			expectedError:  ErrModelInvocation,
			errorCodeCheck: func(err error) bool { return errors.Is(err, ErrModelInvocation) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockBedrockRuntime{
				InvokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					return nil, tc.mockError
				},
			}

			client := NewBedrockClient(mock)
			_, err := client.InvokeClaude(context.Background(), "Test prompt", nil)

			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}

			if !tc.errorCodeCheck(err) {
				t.Errorf("Expected error to wrap %v, got %v", tc.expectedError, err)
			}
		})
	}
}
