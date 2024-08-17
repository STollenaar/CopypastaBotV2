package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)


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
	response := discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintln("Loading..."),
		},
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}
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
		var sqsMessage util.Object
		var destination string
		if interaction.MessageComponentData().CustomID == "command_select" {
			sqsMessage = util.Object{
				Token:         interaction.Token,
				Command:       interaction.MessageComponentData().Values[0],
				GuildID:       interaction.GuildID,
				ApplicationID: interaction.AppID,
				Data:          interaction.MessageComponentData().CustomID,
			}
			destination = "helpReceiver"
		} else {
			sqsMessage = util.Object{
				Token:         interaction.Token,
				Command:       "browse",
				GuildID:       interaction.GuildID,
				ApplicationID: interaction.AppID,
				Data:          interaction.MessageComponentData().CustomID,
			}
			destination = "browseReceiver"
		}
		sqsMessageData, _ := json.Marshal(sqsMessage)
		err := util.PublishObject(destination, string(sqsMessageData))

		if err != nil {
			fmt.Println(err)
		}
	} else {
		cmd := interaction.ApplicationCommandData().Name
		if !util.ContainsCommand(cmd) {
			response = discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command.. How did you even get here?",
				},
			}
		} else {
			// Doing pre-lambda modification
			switch cmd {
			case "caveman":
				fallthrough
			case "respond":
				body, err := extractMessageData("message", interaction)
				if err != nil {
					apiResponse.Body = err.Error()
				}
				req.Body = body
				d, err = json.Marshal(req)
				if err != nil {
					apiResponse.Body = err.Error()
				}
				cmd = "chat"
			case "caveman-vc":
				fallthrough
			case "respond-vc":
				body, err := extractMessageData("chat", interaction)
				if err != nil {
					apiResponse.Body = err.Error()
				}
				req.Body = body
				d, err = json.Marshal(req)
				if err != nil {
					apiResponse.Body = err.Error()
				}
				cmd = "speak"
			}

			// Routing the commands to the correctly lambda that will handle it
			err := util.PublishObject(cmd, string(d))

			if err != nil {
				apiResponse.Body = err.Error()
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

func extractMessageData(optionName string, interaction discordgo.Interaction) (string, error) {
	messageData := interaction.Data.(discordgo.ApplicationCommandInteractionData).Resolved.Messages[interaction.Data.(discordgo.ApplicationCommandInteractionData).TargetID].Content
	appData := interaction.ApplicationCommandData()
	appData.Options = append(appData.Options, &discordgo.ApplicationCommandInteractionDataOption{
		Name:  optionName,
		Type:  3,
		Value: messageData,
	})
	interaction.Data = appData
	iData, err := json.Marshal(interaction)
	if err != nil {
		return "", err
	}
	return string(iData), nil
}
