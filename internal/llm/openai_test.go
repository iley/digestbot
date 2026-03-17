package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIComplete(t *testing.T) {
	var receivedPrompt string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		msgs := req["messages"].([]any)
		msg := msgs[0].(map[string]any)
		receivedPrompt = msg["content"].(string)

		resp := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4.1-mini",
			"choices": []map[string]any{
				{
					"index":         0,
					"finish_reason": "stop",
					"message": map[string]any{
						"role":    "assistant",
						"content": "Test response",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &OpenAI{
		APIKey:  "test-key",
		BaseURL: srv.URL,
	}

	result, err := client.Complete(context.Background(), "Hello, world!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPrompt != "Hello, world!" {
		t.Errorf("prompt not forwarded: got %q, want %q", receivedPrompt, "Hello, world!")
	}

	if result != "Test response" {
		t.Errorf("unexpected result: got %q, want %q", result, "Test response")
	}
}

func TestOpenAICompleteNoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4.1-mini",
			"choices": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &OpenAI{
		APIKey:  "test-key",
		BaseURL: srv.URL,
	}

	_, err := client.Complete(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
}
