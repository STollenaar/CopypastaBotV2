package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stollenaar/copypastabotv2/internal/util"

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

func handler(snsEvent events.SNSEvent) error {
	var req events.APIGatewayProxyRequest
	var interaction discordgo.Interaction
	var response discordgo.WebhookEdit

	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &req)
	if err !=nil {
		return err
	}
	err = json.Unmarshal([]byte(req.Body), &interaction)
	if err !=nil {
		return err
	}


	parsedArguments := util.ParseArguments([]string{"redditpost", "url", "user", "chat"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Redditpost"] == "" && parsedArguments["Url"] == "" && parsedArguments["User"] == "" && parsedArguments["Chat"] == "" {
		response.Content = aws.String("You must provide at least 1 argument")
	} else {
		destination := util.ConfigFile.AWS_SQS_URL
		sqsMessage := util.SQSObject{
			Token:         interaction.Token,
			Command:       interaction.ApplicationCommandData().Name,
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
		}

		if parsedArguments["User"] == "" {

			if parsedArguments["Redditpost"] != "" {
				sqsMessage.Data = parsedArguments["Redditpost"]
				sqsMessage.Type = "redditpost"
				destination = util.ConfigFile.AWS_SQS_URL_OTHER[0]
			} else if parsedArguments["Url"] != "" {
				sqsMessage.Type = "url"
				sqsMessage.Data = parsedArguments["Url"]
			} else if parsedArguments["Chat"] != "" {
				sqsMessage.Type = "chat"
				sqsMessage.Data = parsedArguments["Chat"]
				destination = util.ConfigFile.AWS_SQS_URL_OTHER[1]
			}

			sqsMessageData, _ := json.Marshal(sqsMessage)
			_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
				MessageBody: aws.String(string(sqsMessageData)),
				QueueUrl:    &destination,
			})
			if err != nil {
				fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
				return err
			}
		} else {
			sqsMessage.Type = "user"
			sqsMessage.Data = parsedArguments["User"]
			err := util.ConfigFile.SendStatsBotRequest(sqsMessage)
			
			if err != nil {
				fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
				return err
			}
		}
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)
	return err
}
