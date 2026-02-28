package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

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

// MessageObject general messageobject for functions
type MessageObject struct {
	GuildID   string            `bson:"GuildID" json:"GuildID"`
	ChannelID string            `bson:"ChannelID" json:"ChannelID"`
	MessageID string            `bson:"_id" json:"MessageID"`
	Author    string            `bson:"Author" json:"Author"`
	Content   []string          `bson:"Content" json:"Content"`
	Date      discord.Timestamp `bson:"Date" json:"Date"`
}

type OllamaGenerateResponse struct {
	Model              string    `json:"model"`
	Created            time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context"`
	TotalDuration      int       `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int       `json:"eval_duration"`
}

type OllamaGenerateRequest struct {
	Model  string                 `json:"model"`
	Prompt string                 `json:"prompt"`
	Format map[string]interface{} `json:"format"`
	Stream bool                   `json:"stream"`
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
