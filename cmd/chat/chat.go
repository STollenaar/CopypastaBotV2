package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

var (
	sqsClient *sqs.Client
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Error loading AWS config:", err)
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

	parsedArguments := util.ParseArguments([]string{"message"}, interaction.ApplicationCommandData().Options)

	sqsMessage := util.SQSObject{
		Token:         interaction.Token,
		Command:       interaction.ApplicationCommandData().Name,
		Data:          parsedArguments["Message"],
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
	}

	sqsMessageData, _ := json.Marshal(sqsMessage)
	_, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(string(sqsMessageData)),
		QueueUrl:    aws.String(util.ConfigFile.AWS_SQS_URL),
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)
	return err
}
