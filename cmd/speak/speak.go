package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
)

var (
	sqsClient *sqs.Client
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	sqsClient = sqs.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var interaction discordgo.Interaction
	json.Unmarshal([]byte(req.Body), &interaction)
	response := discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: "Pong",
		},
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	parsedArguments := util.ParseArguments([]string{"redditpost", "url", "user"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Redditpost"] == "" && parsedArguments["Url"] == "" && parsedArguments["User"] == "" {
		response.Data.Content = "You must provide at least 1 argument"
		response.Type = discordgo.InteractionResponseChannelMessageWithSource
	} else {
		destination := util.ConfigFile.AWS_SQS_URL
		sqsMessage := statsUtil.SQSObject{
			Token:         interaction.Token,
			Command:       "speak",
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
		}

		if parsedArguments["Redditpost"] != "" {
			sqsMessage.Data = parsedArguments["Redditpost"]
			sqsMessage.Type = "redditpost"
			destination = util.ConfigFile.AWS_SQS_URL_OTHER
		} else if parsedArguments["Url"] != "" {
			sqsMessage.Type = "url"
			sqsMessage.Data = parsedArguments["Url"]
		} else {
			sqsMessage.Type = "user"
			sqsMessage.Data = parsedArguments["User"]
		}

		sqsMessageData, _ := json.Marshal(sqsMessage)
		_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(sqsMessageData)),
			QueueUrl:    &destination,
		})
		if err != nil {
			fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
		}
	}
	data, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}
