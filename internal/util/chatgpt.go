package util

import (
	"context"
	_ "embed"
	"log"
	"strings"

	"github.com/ayush6624/go-chatgpt"
)

var (
	chatGPTClient *chatgpt.Client

	//go:embed wrapSSML.txt
	wrapPrompt string
)

func initChatGPT() {
	openAIKey, err := ConfigFile.GetOpenAIKey()
	if err != nil {
		log.Fatal(err)
	}

	chatGPTClient, err = chatgpt.NewClient(openAIKey)
	if err != nil {
		log.Fatal(err)
	}
}

func GetChatGPTResponse(systemPrompt, userInput, userID string) (*chatgpt.ChatResponse, error) {
	if chatGPTClient == nil {
		initChatGPT()
	}
	return chatGPTClient.Send(context.TODO(), &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT35Turbo,
		User:  userID,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    chatgpt.ChatGPTModelRoleUser,
				Content: userInput,
			},
		},
	})
}

func WrapIntoSSML(input, userID string) (*chatgpt.ChatResponse, error) {
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
