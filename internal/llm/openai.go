package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const defaultModel = "gpt-5-mini"

// OpenAI implements LLM using the OpenAI chat completions API.
type OpenAI struct {
	APIKey  string
	Model   string
	BaseURL string // OpenAI-compatible endpoint; leave empty for OpenAI
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

	return cleanResponse(completion.Choices[0].Message.Content), nil
}

var thinkBlockRe = regexp.MustCompile(`(?s)<think>.*?</think>`)

// cleanResponse strips artifacts that some models wrap around their output so
// callers get the bare content. Reasoning models (e.g. DeepSeek-R1, Qwen3 in
// thinking mode) prepend a <think>...</think> block, and many models fence
// structured output in a ```json ... ``` block; both break strict JSON parsing.
func cleanResponse(s string) string {
	s = thinkBlockRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	// Strip a surrounding markdown code fence (```json ... ``` or ``` ... ```).
	if strings.HasPrefix(s, "```") {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}

	return strings.TrimSpace(s)
}
