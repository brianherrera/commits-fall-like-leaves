package bedrock

type ClaudeRequest struct {
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ContentBlock struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type Message struct {
	Content []ContentBlock `json:"content"`
	Role    string         `json:"role"`
}

type ClaudeResponse struct {
	Content []ContentBlock `json:"content"`
}

type ClaudeOptions struct {
	MaxTokens   int     // Maximum number of tokens to generate (default: 500)
	Temperature float64 // Controls randomness (0.0-1.0, default: 0.7)
	System      string  // Defines the bounds of your task’s specific requirements.
}

func DefaultClaudeOptions() ClaudeOptions {
	return ClaudeOptions{
		MaxTokens:   DefaultMaxTokens,
		Temperature: DefaultTemperature,
	}
}
