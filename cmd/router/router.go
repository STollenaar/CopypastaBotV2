package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	lambdaService "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	lambdaClient *lambdaService.Client
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Error loading AWS config:", err)
	}
	lambdaClient = lambdaService.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	d, _ := json.Marshal(req)
	fmt.Println(string(d))
	// Doing the body and signature verification
	verified := util.IsVerified(req.Body, req.Headers["x-signature-ed25519"], req.Headers["x-signature-timestamp"])
	if !verified {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       string("Invalid request signature"),
		}, nil
	}

	var interaction discordgo.Interaction

	json.Unmarshal([]byte(req.Body), &interaction)

	if interaction.Type == discordgo.InteractionType(discordgo.InteractionResponsePong) {
		response := util.ResponseObject{
			Type: discordgo.InteractionResponsePong,
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

	// Routing the commands to the correctly lambda that will handle it
	switch interaction.ApplicationCommandData().Name {
	case "ping":
		out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
			FunctionName: aws.String("ping"),
			Payload:      d,
		})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: err.Error(),
			}, nil
		}
		var response events.APIGatewayProxyResponse
		json.Unmarshal(out.Payload, &response)
		return response, nil
	case "pasta":
		out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
			FunctionName: aws.String("pasta"),
			Payload:      d,
		})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: err.Error(),
			}, nil
		}
		var response events.APIGatewayProxyResponse
		json.Unmarshal(out.Payload, &response)
		return response, nil
	case "markov":
		out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
			FunctionName: aws.String("markov"),
			Payload:      d,
		})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: err.Error(),
			}, nil
		}
		var response events.APIGatewayProxyResponse
		json.Unmarshal(out.Payload, &response)
		return response, nil
	default:
		response := util.ResponseObject{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: discordgo.InteractionResponseData{
				Content: "Unknown command.. How did you even get here?",
			},
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
}
