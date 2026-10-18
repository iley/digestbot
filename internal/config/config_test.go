package config

import (
	"os"
	"strings"
	"testing"
)

func TestNewsRankerConfig(t *testing.T) {
	for _, tc := range []struct {
		name       string
		env        map[string]string
		args       []string
		wantRanker string
		wantErr    string
	}{
		{name: "default", wantRanker: "llm"},
		{name: "jev env", env: map[string]string{"DIGESTBOT_NEWS_RANKER": "jev", "DIGESTBOT_JEV_API_KEY": "env-key"}, wantRanker: "jev"},
		{name: "flags override env", env: map[string]string{"DIGESTBOT_NEWS_RANKER": "llm", "DIGESTBOT_JEV_API_KEY": "env-key", "DIGESTBOT_JEV_MODEL": "env-model", "DIGESTBOT_JEV_BASE_URL": "env-url"}, args: []string{"--news-ranker=jev", "--jev-api-key=flag-key", "--jev-model=flag-model", "--jev-base-url=https://example.com/v1"}, wantRanker: "jev"},
		{name: "missing jev key", args: []string{"--news-ranker=jev"}, wantErr: "Jev API key is required"},
		{name: "blank jev key", args: []string{"--news-ranker=jev", "--jev-api-key= "}, wantErr: "Jev API key is required"},
		{name: "invalid ranker", args: []string{"--news-ranker=unknown"}, wantErr: "invalid news ranker"},
		{name: "summarizer still needs key", args: []string{"--news-ranker=jev", "--jev-api-key=key", "--openai-api-key="}, wantErr: "LLM API key is required"},
		{name: "weather needs no model keys", args: []string{"--segments=weather", "--latitude=0", "--longitude=0", "--news-ranker=jev", "--openai-api-key="}, wantRanker: "jev"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"BOT_TOKEN", "CHAT_ID", "LATITUDE", "LONGITUDE", "TIMEZONE", "OPENAI_API_KEY", "LLM_BASE_URL", "LLM_MODEL", "SEGMENTS", "NEWS_RANKER", "JEV_API_KEY", "JEV_MODEL", "JEV_BASE_URL"} {
				t.Setenv("DIGESTBOT_"+key, "")
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })
			os.Args = append([]string{"digestbot", "--bot-token=test", "--chat-id=1", "--segments=meduza", "--openai-api-key=llm-key"}, tc.args...)
			cfg, err := Parse()
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.NewsRanker != tc.wantRanker {
				t.Fatalf("ranker %q", cfg.NewsRanker)
			}
			if tc.name == "jev env" && cfg.JevAPIKey != "env-key" {
				t.Fatal("missing environment key")
			}
			if tc.name == "flags override env" && (cfg.JevAPIKey != "flag-key" || cfg.JevModel != "flag-model" || cfg.JevBaseURL != "https://example.com/v1") {
				t.Fatal("flags did not override environment")
			}
		})
	}
}
