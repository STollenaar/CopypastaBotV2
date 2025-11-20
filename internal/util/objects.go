package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/bwmarrin/discordgo"
)

var (
	snsClient *sns.Client
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	snsClient = sns.NewFromConfig(cfg)
}

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
	GuildID   string               `bson:"GuildID" json:"GuildID"`
	ChannelID string               `bson:"ChannelID" json:"ChannelID"`
	MessageID string               `bson:"_id" json:"MessageID"`
	Author    string               `bson:"Author" json:"Author"`
	Content   []string             `bson:"Content" json:"Content"`
	Date      discordgo.TimeStamps `bson:"Date" json:"Date"`
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
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format map[string]interface{} `json:"format"`
	Stream bool   `json:"stream"`
}

func GetMessageObject(object Object) (discordgo.Message, error) {
	if _, err := strconv.Atoi(object.Token); err == nil {
		return discordgo.Message{}, errors.New("object token is a snowflake")
	}
	resp, err := SendRequest("GET", object.ApplicationID, object.Token, WEBHOOK, []byte{})
	var bodyString string
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString = string(bodyData)
		fmt.Println(resp, bodyString)
	}
	if err != nil {
		return discordgo.Message{}, fmt.Errorf("error sending request %v", err)
	}
	var message discordgo.Message
	err = json.Unmarshal([]byte(bodyString), &message)
	if err != nil {
		return discordgo.Message{}, fmt.Errorf("error parsing into interaction with data: %s, and error: %v", bodyString, err)
	}
	return message, nil
}

func Pointer[T any](d T) *T {
	return &d
}