package llm

import "context"

// LLM provides text completion from a language model.
type LLM interface {
	Complete(ctx context.Context, prompt string) (string, error)
}
