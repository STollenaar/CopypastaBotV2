package main

import (
	"encoding/json"
	"fmt"

	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bwmarrin/discordgo"
)

func main() {
	lambda.Start(handler)
}

func handler(snsEvent events.SNSEvent) error {
	var req events.APIGatewayProxyRequest
	var interaction discordgo.Interaction
	var response discordgo.WebhookEdit

	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &req)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(req.Body), &interaction)
	if err != nil {
		return err
	}

	parsedArguments := util.ParseArguments([]string{"redditpost", "url", "user", "chat"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Redditpost"] == "" && parsedArguments["Url"] == "" && parsedArguments["User"] == "" && parsedArguments["Chat"] == "" {
		response.Content = aws.String("You must provide at least 1 argument")
	} else {
		sqsMessage := util.Object{
			Token:         interaction.Token,
			Command:       interaction.ApplicationCommandData().Name,
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
		}

		if parsedArguments["User"] == "" {
			destination := "sqsReceiver"
			// destination := util.ConfigFile.AWS_SQS_URL

			if parsedArguments["Redditpost"] != "" {
				sqsMessage.Data = parsedArguments["Redditpost"]
				sqsMessage.Type = "redditpost"
				destination = "speakReceiver"
			} else if parsedArguments["Url"] != "" {
				sqsMessage.Type = "url"
				sqsMessage.Data = parsedArguments["Url"]
			} else if parsedArguments["Chat"] != "" {
				sqsMessage.Type = "chat"
				sqsMessage.Data = parsedArguments["Chat"]
				destination = "chatReceiver"
			}

			sqsMessageData, _ := json.Marshal(sqsMessage)
			err := util.PublishObject(destination, string(sqsMessageData))
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
