package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const defaultModel = "gpt-5-mini"

// OpenAI implements LLM using the OpenAI chat completions API.
type OpenAI struct {
	APIKey  string
	Model   string
	BaseURL string // override for testing; leave empty for default
}

func (o *OpenAI) Complete(ctx context.Context, prompt string) (string, error) {
	model := o.Model
	if model == "" {
		model = defaultModel
	}

	opts := []option.RequestOption{option.WithAPIKey(o.APIKey)}
	if o.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(o.BaseURL))
	}
	client := openai.NewClient(opts...)

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai completion: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices in response")
	}

	return completion.Choices[0].Message.Content, nil
}
