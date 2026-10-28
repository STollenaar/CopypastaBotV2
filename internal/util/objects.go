package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/disgoorg/disgo/discord"
)

type Object struct {
	Type          string `json:"type"`
	Command       string `json:"command"`
	Data          string `json:"data"`
	ChannelID     string `json:"channelID"`
	GuildID       string `json:"guildID"`
	Token         string `json:"token"`
	ApplicationID string `json:"applicationID"`
}

// UserMessagesResponse is the statistics bot's reply to /userMessages. Data
// holds one entry per message rather than a single joined string: a markov
// chain needs to know where each message ended, otherwise it runs from the tail
// of one message straight into the head of the next.
type UserMessagesResponse struct {
	Type          string   `json:"type"`
	Command       string   `json:"command"`
	Data          []string `json:"data"`
	ChannelID     string   `json:"channelID"`
	GuildID       string   `json:"guildID"`
	Token         string   `json:"token"`
	ApplicationID string   `json:"applicationID"`
}

// MessageObject general messageobject for functions
type MessageObject struct {
	GuildID   string            `bson:"GuildID" json:"GuildID"`
	ChannelID string            `bson:"ChannelID" json:"ChannelID"`
	MessageID string            `bson:"_id" json:"MessageID"`
	Author    string            `bson:"Author" json:"Author"`
	Content   []string          `bson:"Content" json:"Content"`
	Date      discord.Timestamp `bson:"Date" json:"Date"`
}

// ChatCompletionRequest is a LiteLLM (OpenAI-compatible) /chat/completions request.
type ChatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
	Stream   bool                    `json:"stream"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                   `json:"index"`
		Message      ChatCompletionMessage `json:"message"`
		FinishReason string                `json:"finish_reason"`
	} `json:"choices"`
}

func GetMessageObject(object Object) (discord.Message, error) {
	if _, err := strconv.Atoi(object.Token); err == nil {
		return discord.Message{}, errors.New("object token is a snowflake")
	}
	resp, err := SendRequest("GET", object.ApplicationID, object.Token, WEBHOOK, []byte{})
	var bodyString string
	if resp != nil {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			slog.Error("Error reading body", slog.Any("err", err))
		}

		bodyData := buf.String()

		bodyString = bodyData
		slog.Debug("HTTP response", slog.Any("response", resp), slog.String("body", bodyString))
	}
	if err != nil {
		return discord.Message{}, fmt.Errorf("error sending request %v", err)
	}
	var message discord.Message
	err = json.Unmarshal([]byte(bodyString), &message)
	if err != nil {
		return discord.Message{}, fmt.Errorf("error parsing into interaction with data: %s, and error: %v", bodyString, err)
	}
	return message, nil
}

func Pointer[T any](d T) *T {
	return &d
}
