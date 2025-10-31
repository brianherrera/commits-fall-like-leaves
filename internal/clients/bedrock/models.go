package bedrock

type ClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

type ClaudeOptions struct {
	MaxTokens     int      // Maximum number of tokens to generate (default: 500)
	Temperature   float64  // Controls randomness (0.0-1.0, default: 0.7)
	TopP          float64  // Controls diversity via nucleus sampling (0.0-1.0, default: 0.9)
	TopK          int      // Controls diversity via top-k sampling (default: 0)
	StopSequences []string // Sequences where the API will stop generating further tokens
}

func DefaultClaudeOptions() ClaudeOptions {
	return ClaudeOptions{
		MaxTokens:   DefaultMaxTokens,
		Temperature: DefaultTemperature,
		TopP:        DefaultTopP,
		TopK:        DefaultTopK,
	}
}
