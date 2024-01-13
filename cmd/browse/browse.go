package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"

	"github.com/bwmarrin/discordgo"
)

var (
	sqsClient      *sqs.Client
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

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintln("Loading..."),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	var interaction discordgo.Interaction
	json.Unmarshal([]byte(req.Body), &interaction)

	parsedArguments := util.ParseArguments([]string{"name"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Name"] == "" {
		parsedArguments["Name"] = "all"
	}

	sqsMessage := statsUtil.SQSObject{
		Token:         interaction.Token,
		Command:       "browse",
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
		Data:          parsedArguments["Name"],
	}
	sqsMessageData, _ := json.Marshal(sqsMessage)
	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(string(sqsMessageData)),
		QueueUrl:    aws.String(util.ConfigFile.AWS_SQS_URL),
	})
	if err != nil {
		fmt.Println(err)
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}
