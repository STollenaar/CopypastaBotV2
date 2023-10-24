package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
)

func main() {
	lambda.Start(handler)
}

func handler() (events.APIGatewayProxyResponse, error) {
	response := discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: "Pong",
		},
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("error marshalling response data: %w", err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}
