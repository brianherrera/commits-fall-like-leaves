package bedrock

const (
	AnthropicVersion = "bedrock-2023-05-31"
	ClaudeModelID    = "global.anthropic.claude-haiku-4-5-20251001-v1:0" // Obviously.

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
