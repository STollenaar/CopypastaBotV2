package util

import (
	"context"
	_ "embed"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var (
	chatGPTClient *openai.Client

	//go:embed wrapSSML.txt
	wrapPrompt string
)

func initChatGPT() {
	openAIKey, err := ConfigFile.GetOpenAIKey()
	if err != nil {
		log.Fatal(err)
	}

	chatGPTClient = openai.NewClient(openAIKey)
}

func GetChatGPTResponse(systemPrompt, userInput, userID string) (openai.ChatCompletionResponse, error) {
	if chatGPTClient == nil {
		initChatGPT()
	}

	return chatGPTClient.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		User:  userID,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userInput,
			},
		},
	})
}

func WrapIntoSSML(input, userID string) (openai.ChatCompletionResponse, error) {
	return GetChatGPTResponse(wrapPrompt, escapeSSML(input), userID)
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
