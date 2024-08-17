package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
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

func PublishObject(destination, data string) error {
	// Routing the commands to the correctly lambda that will handle it
	messageAttributes := make(map[string]types.MessageAttributeValue)
	messageAttributes["function_name"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(destination),
	}

	_, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn:          &ConfigFile.AWS_SNS_TOPIC_ARN,
		Message:           aws.String(string(data)),
		MessageAttributes: messageAttributes,
	})
	return err
}
