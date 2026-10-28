package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var thinkBlock = regexp.MustCompile(`(?s)^\s*<think>.*?</think>\s*`)

var llmClient = &http.Client{Timeout: 30 * time.Minute}

// CreateLLMCompletion sends a system prompt and user message to LiteLLM's
// /chat/completions endpoint and returns the reply text.
func CreateLLMCompletion(systemPrompt, message string) (string, error) {
	data, err := json.Marshal(ChatCompletionRequest{
		Model: ConfigFile.LITELLM_MODEL,
		Messages: []ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", liteLLMURL(), bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	token, err := GetLiteLLMAPIKey()
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := llmClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("litellm request returned non-200", slog.Int("status", resp.StatusCode), slog.String("body", buf.String()))
		return "", errors.New(buf.String())
	}

	var r ChatCompletionResponse
	if err := json.Unmarshal(buf.Bytes(), &r); err != nil {
		return "", err
	}
	if len(r.Choices) == 0 {
		return "", errors.New("litellm returned no choices")
	}
	// Reasoning models can leak their thinking into the content.
	return thinkBlock.ReplaceAllString(r.Choices[0].Message.Content, ""), nil
}

// liteLLMURL returns LITELLM_URL, defaulting to http:// when no scheme is given.
func liteLLMURL() string {
	url := ConfigFile.LITELLM_URL
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return fmt.Sprintf("http://%s", url)
}
