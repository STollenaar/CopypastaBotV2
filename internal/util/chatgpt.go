package util

import (
	"context"
	_ "embed"
	"log"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const claudeModel = "claude-sonnet-4-6"

var (
	claudeClient *anthropic.Client

	//go:embed wrapSSML.txt
	wrapPrompt string
)

func initClaude() {
	apiKey, err := ConfigFile.GetAnthropicKey()
	if err != nil {
		log.Fatal(err)
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	claudeClient = &c
}

func GetClaudeResponse(systemPrompt, userInput string) (string, error) {
	if claudeClient == nil {
		initClaude()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	msg, err := claudeClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     claudeModel,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
		},
	},
)
	if err != nil {
		return "", err
	}
	if len(msg.Content) == 0 {
		return "", nil
	}
	return msg.Content[0].Text, nil
}

func WrapIntoSSML(input, _ string) (string, error) {
	return GetClaudeResponse(wrapPrompt, escapeSSML(input))
}

func escapeSSML(input string) string {
	if strings.Contains(input, "'") {
		input = strings.ReplaceAll(input, "'", "&apos;")
	}
	if strings.Contains(input, `"`) {
		input = strings.ReplaceAll(input, `"`, "&quot;")
	}
	if strings.Contains(input, "&") {
		input = strings.ReplaceAll(input, "&", "&amp;")
	}
	if strings.Contains(input, "<") {
		input = strings.ReplaceAll(input, "<", "&lt;")
	}
	if strings.Contains(input, ">") {
		input = strings.ReplaceAll(input, ">", "&gt;")
	}
	return input
}
