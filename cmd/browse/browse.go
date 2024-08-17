package main

import (
	"encoding/json"
	"fmt"

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

	parsedArguments := util.ParseArguments([]string{"name"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Name"] == "" {
		parsedArguments["Name"] = "all"
	}

	sqsMessage := util.Object{
		Token:         interaction.Token,
		Command:       "browse",
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
		Data:          parsedArguments["Name"],
	}
	sqsMessageData, _ := json.Marshal(sqsMessage)
	err = util.PublishObject("browseReceiver", string(sqsMessageData))
	if err != nil {
		fmt.Printf("Encountered an error while processing the browse command: %v\n", err)
		return err
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)
	return err
}
