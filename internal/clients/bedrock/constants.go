package bedrock

const (
	ClaudeModelID = "anthropic.claude-3-5-haiku-20241022-v1:0" // Obviously.

	DefaultMaxTokens   = 500
	DefaultTemperature = 0.7

	// AWS Bedrock error codes
	ValidationExceptionCode           = "ValidationException"
	ResourceNotFoundExceptionCode     = "ResourceNotFoundException"
	ThrottlingExceptionCode           = "ThrottlingException"
	ServiceQuotaExceededExceptionCode = "ServiceQuotaExceededException"
	AccessDeniedExceptionCode         = "AccessDeniedException"
	InternalServerExceptionCode       = "InternalServerException"
)
