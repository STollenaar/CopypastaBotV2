package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

func main() {
	lambda.Start(handler)
}

func handler(snsEvent events.SNSEvent) error {
	response := discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintln("Loading..."),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	var interaction discordgo.Interaction
	json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &interaction)

	sqsMessage := util.Object{
		Token:         interaction.Token,
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
	}

	sqsMessageData, _ := json.Marshal(sqsMessage)
	err := util.PublishObject("helpReceiver", string(sqsMessageData))
	if err != nil {
		fmt.Printf("Encountered an error while processing the help command: %v\n", err)
		return err
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Responding with %s\n", string(data))
	return nil
}
