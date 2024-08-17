package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stollenaar/copypastabotv2/internal/util"

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

	sqsMessage := util.Object{
		Token:         interaction.Token,
		Command:       "markov",
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
	}

	parsedArguments := util.ParseArguments([]string{"url", "user"}, interaction.ApplicationCommandData().Options)

	if url, ok := parsedArguments["Url"]; ok {
		sqsMessage.Type = "url"
		sqsMessage.Data = url

		sqsMessageData, _ := json.Marshal(sqsMessage)
		err := util.PublishObject("sqsReceiver", string(sqsMessageData))
		if err != nil {
			fmt.Printf("Encountered an error while processing the markov command: %v\n", err)
			return err
		}
	} else if user, ok := parsedArguments["User"]; ok {
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")
		sqsMessage.Type = "user"
		sqsMessage.Data = user

		err := util.ConfigFile.SendStatsBotRequest(sqsMessage)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)

	return err
}
