package haiku

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/brianherrera/commits-fall-like-leaves/internal/clients/bedrock"
)

var (
	ErrBadHaikuRequest = errors.New("bad haiku request received")
	ErrCreateHaiku     = errors.New("error creating commit message haiku")
)

type BedrockClient interface {
	InvokeClaude(ctx context.Context, prompt string, opts *bedrock.ClaudeOptions) (string, error)
}

type HaikuService struct {
	bedrockClient BedrockClient
}

func NewHaikuService(bedrockClient BedrockClient) *HaikuService {
	return &HaikuService{
		bedrockClient: bedrockClient,
	}
}

func NewDefaultHaikuService(cfg aws.Config) *HaikuService {
	return NewHaikuService(bedrock.NewDefaultBedrockClient(cfg))
}

func (h *HaikuService) CreateHaiku(ctx context.Context, request HaikuCommitRequest) (HaikuCommitResponse, error) {
	mood := request.Mood
	if mood != "" && !mood.IsValid() {
		log.Printf("[HAIKU SERVICE] invalid mood: %s\n", mood)
		return HaikuCommitResponse{}, ErrBadHaikuRequest
	}

	if request.Mood == "" {
		mood = MoodReflective
	}

	prompt := fmt.Sprintf("Create a %s haiku from this commit message: %s", mood, request.CommitMessage)

	options := &bedrock.ClaudeOptions{
		System: HaikuSystemPrompt,
	}

	log.Printf("[HAIKU SERVICE] sending request to Bedrock: %s\n", prompt)
	response, err := h.bedrockClient.InvokeClaude(ctx, prompt, options)
	if err != nil {
		log.Printf("[HAIKU SERVICE] error invoking Claude: %v\n", err)
		return HaikuCommitResponse{}, fmt.Errorf("%w: invoking Claude: %v", ErrCreateHaiku, err)
	}

	return HaikuCommitResponse{
		Haiku: response,
	}, nil
}
