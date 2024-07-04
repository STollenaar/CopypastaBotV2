package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	//go:embed respond.txt
	systemPrompt string

	sqsObject util.SQSObject
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), &sqsObject)
	if err != nil {
		fmt.Println(err)
		return err
	}

	chatRSP, err := util.GetChatGPTResponse(systemPrompt, sqsObject.Data, sqsObject.Type)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Getting around the 4096 word limit
	contents := util.BreakContent("<@"+sqsObject.Type+"> "+chatRSP.Choices[0].Message.Content, 4096)
	var embeds []*discordgo.MessageEmbed
	for _, content := range contents {
		embed := discordgo.MessageEmbed{}
		embed.Description = content
		embeds = append(embeds, &embed)
	}
	author := strings.Split(sqsObject.Token, ";")
	response := discordgo.WebhookParams{
		Embeds:    embeds,
		Flags:     discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsUrgent,
		Username:  author[0],
		AvatarURL: author[1],
		// ENABLE WHEN READY
		// Reference: &discordgo.MessageReference{
		// 	MessageID: sqsObject.Token,
		// 	ChannelID: sqsObject.ChannelID,
		// 	GuildID:   sqsObject.GuildID,
		// },
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Users: []string{sqsObject.Type},
			Roles: []string{},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return err
	}
	resp, err := util.SendAsWebhook(data)
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString := string(bodyData)
		fmt.Println(resp, bodyString)
	}

	return err
}
