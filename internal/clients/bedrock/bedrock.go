// Package clients provides client implementations for various external services.
//
// This package contains client implementations for AWS Bedrock and other services
// that the application depends on. Each client is designed to be easily configurable
// and to provide a clean, idiomatic Go interface to the underlying service.
package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockRuntime interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

type BedrockClient struct {
	runtimeClient BedrockRuntime
}

func NewBedrockClient(runtimeClient BedrockRuntime) *BedrockClient {
	return &BedrockClient{
		runtimeClient: runtimeClient,
	}
}

func NewDefaultBedrockClient(cfg aws.Config) *BedrockClient {
	return NewBedrockClient(bedrockruntime.NewFromConfig(cfg))
}

func (c *BedrockClient) InvokeClaude(ctx context.Context, prompt string, opts *ClaudeOptions) (string, error) {
	// Validate prompt
	if prompt == "" {
		log.Printf("[BEDROCK CLIENT] prompt is empty")
		return "", fmt.Errorf("%w: prompt cannot be empty", ErrInvalidRequest)
	}

	modelID := ClaudeModelID

	options := DefaultClaudeOptions()
	if opts != nil {
		// Validate and apply MaxTokens (must be positive)
		if opts.MaxTokens > 0 {
			options.MaxTokens = opts.MaxTokens
		}
		// Validate and apply Temperature (must be between 0.0 and 1.0)
		if opts.Temperature > 0 && opts.Temperature <= 1.0 {
			options.Temperature = opts.Temperature
		}
		if opts.System != "" {
			options.System = opts.System
		}
	}

	request := &ClaudeRequest{
		MaxTokens: options.MaxTokens,
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentBlock{
					{
						Text: prompt,
						Type: "text",
					},
				},
			},
		},
		Temperature: options.Temperature,
		System:      options.System,
	}

	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("[BEDROCK CLIENT] error encountered marshalling request: %v", err)
		return "", fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	output, err := c.runtimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		log.Printf("[BEDROCK CLIENT] error encountered invoking model: %v", err)
		return "", c.handleBedrockError(err)
	}

	var response ClaudeResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		log.Printf("[BEDROCK CLIENT] error encountered parsing response: %v", err)
		return "", fmt.Errorf("%w: %v", ErrResponseParsing, err)
	}

	return response.Content[0].Text, nil
}
