package bedrock

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrModelInvocation  = errors.New("model invocation failed")
	ErrResponseParsing  = errors.New("failed to parse model response")
	ErrModelUnavailable = errors.New("model is currently unavailable")
	ErrQuotaExceeded    = errors.New("quota exceeded for model invocation")
	ErrThrottling       = errors.New("request was throttled")
	ErrValidation       = errors.New("validation error")
)

// BedrockError represents a detailed error from Bedrock API
type BedrockError struct {
	OriginalErr error
	Message     string
	StatusCode  int
	RequestID   string
}

func (e *BedrockError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("bedrock error (status: %d, request-id: %s): %s",
			e.StatusCode, e.RequestID, e.Message)
	}
	return fmt.Sprintf("bedrock error: %s", e.Message)
}

func (e *BedrockError) Unwrap() error {
	return e.OriginalErr
}

// handleBedrockError processes AWS errors and returns a more specific error.
func (c *BedrockClient) handleBedrockError(err error) error {
	var bedrockErr *BedrockError

	// Extract AWS error details
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		bedrockErr = &BedrockError{
			OriginalErr: err,
			Message:     awsErr.Error(),
			RequestID:   awsErr.ErrorCode(),
		}

		var respErr *smithyhttp.ResponseError
		if errors.As(err, &respErr) {
			if respErr.Response != nil {
				bedrockErr.StatusCode = respErr.Response.StatusCode
			}
		}

		// Map to specific error types using constants instead of string checks
		errorCode := awsErr.ErrorCode()
		switch {
		case strings.Contains(errorCode, ValidationExceptionCode):
			return fmt.Errorf("%w: %v", ErrValidation, bedrockErr)
		case strings.Contains(errorCode, ResourceNotFoundExceptionCode):
			return fmt.Errorf("%w: %v", ErrModelUnavailable, bedrockErr)
		case strings.Contains(errorCode, ThrottlingExceptionCode):
			return fmt.Errorf("%w: %v", ErrThrottling, bedrockErr)
		case strings.Contains(errorCode, ServiceQuotaExceededExceptionCode):
			return fmt.Errorf("%w: %v", ErrQuotaExceeded, bedrockErr)
		case strings.Contains(errorCode, AccessDeniedExceptionCode):
			return fmt.Errorf("%w: %v", ErrValidation, bedrockErr)
		case strings.Contains(errorCode, InternalServerExceptionCode):
			return fmt.Errorf("%w: %v", ErrModelInvocation, bedrockErr)
		default:
			return fmt.Errorf("%w: %v", ErrModelInvocation, bedrockErr)
		}
	}

	// If not an AWS error, wrap with generic error
	return fmt.Errorf("%w: %v", ErrModelInvocation, err)
}
