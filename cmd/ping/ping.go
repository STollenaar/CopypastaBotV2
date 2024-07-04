package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

func main() {
	lambda.Start(handler)
}

func handler(snsEvent events.SNSEvent) error {
	var req events.APIGatewayProxyRequest
	var interaction discordgo.Interaction
	response := discordgo.WebhookEdit{
		Content: aws.String("Pong"),
	}

	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &req)
	if err !=nil {
		return err
	}
	err = json.Unmarshal([]byte(req.Body), &interaction)
	if err !=nil {
		return err
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)
	return err
}
