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
			response := ClaudeResponse{Completion: "Test response"}
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

func TestInvokeClaudeSuccess(t *testing.T) {
	expectedCompletion := "This is a test response from Claude"

	mock := &MockBedrockRuntime{
		InvokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
			// Verify request parameters
			if params.ModelId == nil || *params.ModelId != ClaudeModelID {
				t.Errorf("Expected model ID %s, got %v", ClaudeModelID, params.ModelId)
			}

			if params.ContentType == nil || *params.ContentType != "application/json" {
				t.Errorf("Expected content type application/json, got %v", params.ContentType)
			}

			// Parse the request body to verify parameters
			var request ClaudeRequest
			if err := json.Unmarshal(params.Body, &request); err != nil {
				t.Errorf("Failed to unmarshal request body: %v", err)
			}

			// Return mock response
			response := ClaudeResponse{Completion: expectedCompletion}
			responseBytes, _ := json.Marshal(response)
			return &bedrockruntime.InvokeModelOutput{
				Body: responseBytes,
			}, nil
		},
	}

	client := NewBedrockClient(mock)

	completion, err := client.InvokeClaude(context.Background(), "Test prompt", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if completion != expectedCompletion {
		t.Errorf("Expected completion %q, got %q", expectedCompletion, completion)
	}

	// Test with custom options
	customOpts := &ClaudeOptions{
		MaxTokens:   200,
		Temperature: 0.5,
	}

	completion, err = client.InvokeClaude(context.Background(), "Test prompt with options", customOpts)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if completion != expectedCompletion {
		t.Errorf("Expected completion %q, got %q", expectedCompletion, completion)
	}
}

// TestInvokeClaudeResponseParsingError tests the case when response parsing fails
func TestInvokeClaudeResponseParsingError(t *testing.T) {
	mock := &MockBedrockRuntime{
		InvokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
			// Return invalid JSON that will cause unmarshaling to fail
			return &bedrockruntime.InvokeModelOutput{
				Body: []byte(`{"invalid_json": `),
			}, nil
		},
	}

	client := NewBedrockClient(mock)
	_, err := client.InvokeClaude(context.Background(), "Test prompt", nil)

	if err == nil {
		t.Errorf("Expected error but got nil")
		return
	}

	if !errors.Is(err, ErrResponseParsing) {
		t.Errorf("Expected error to wrap ErrResponseParsing, got %v", err)
	}

	expectedErrText := "failed to parse model response"
	if !strings.Contains(err.Error(), expectedErrText) {
		t.Errorf("Expected error message to contain %q, got %q", expectedErrText, err.Error())
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
