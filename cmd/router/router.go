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
	lambdaService "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

var (
	lambdaClient *lambdaService.Client
	sqsClient    *sqs.Client
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Error loading AWS config:", err)
	}
	lambdaClient = lambdaService.NewFromConfig(cfg)
	sqsClient = sqs.NewFromConfig(cfg)
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
	var response discordgo.InteractionResponse
	apiResponse := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	json.Unmarshal([]byte(req.Body), &interaction)

	if interaction.Type == discordgo.InteractionType(discordgo.InteractionResponsePong) {
		response.Type = discordgo.InteractionResponsePong
	} else if interaction.Type == discordgo.InteractionMessageComponent {
		response.Type = discordgo.InteractionResponseDeferredMessageUpdate

		sqsMessage := statsUtil.SQSObject{
			Token:         interaction.Token,
			Command:       "browse",
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
			Data:          interaction.MessageComponentData().CustomID,
		}
		sqsMessageData, _ := json.Marshal(sqsMessage)
		_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody:  aws.String(string(sqsMessageData)),
			QueueUrl:     aws.String(util.ConfigFile.AWS_SQS_URL),
			DelaySeconds: *aws.Int32(2),
		})
		if err != nil {
			fmt.Println(err)
		}
	} else {
		// Routing the commands to the correctly lambda that will handle it
		switch interaction.ApplicationCommandData().Name {
		case "browse":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("browse"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		case "chat":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("chat"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		case "markov":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("markov"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		case "ping":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("ping"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		case "pasta":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("pasta"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		case "speak":
			out, err := lambdaClient.Invoke(context.TODO(), &lambdaService.InvokeInput{
				FunctionName: aws.String("speak"),
				Payload:      d,
			})
			if err != nil {
				apiResponse.Body = err.Error()
			} else {
				json.Unmarshal(out.Payload, &apiResponse)
			}
		default:
			response = discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command.. How did you even get here?",
				},
			}
		}
	}

	if apiResponse.Body == "" {
		data, _ := json.Marshal(response)
		apiResponse.Body = string(data)
	}
	fmt.Printf("APIResponse: %v\n", apiResponse)
	return apiResponse, nil
}
