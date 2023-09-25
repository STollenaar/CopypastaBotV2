package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"

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

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := util.ResponseObject{
		Data: discordgo.InteractionResponseData{
			Content: fmt.Sprintln("Loading..."),
		},
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	var interaction discordgo.Interaction
	json.Unmarshal([]byte(req.Body), &interaction)

	sqsMessage := statsUtil.SQSObject{
		Token:         interaction.Token,
		Command:       "markov",
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
	}

	parsedArguments := util.ParseArguments([]string{"url", "user"}, interaction.ApplicationCommandData().Options)

	if url, ok := parsedArguments["Url"]; ok {
		sqsMessage.Type = "url"
		sqsMessage.Data = url
	} else if user, ok := parsedArguments["User"]; ok {
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")
		sqsMessage.Type = "user"
		sqsMessage.Data = user
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
