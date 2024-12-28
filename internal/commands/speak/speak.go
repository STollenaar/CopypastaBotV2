package speak

import (
	"fmt"

	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bwmarrin/discordgo"
)

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})

	var response discordgo.WebhookEdit

	parsedArguments := util.ParseArguments([]string{"redditpost", "url", "user", "chat"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Redditpost"] == "" && parsedArguments["Url"] == "" && parsedArguments["User"] == "" && parsedArguments["Chat"] == "" {
		response.Content = aws.String("You must provide at least 1 argument")
		bot.InteractionResponseEdit(interaction.Interaction, &response)
	} else {
		sqsMessage := util.Object{
			Token:         interaction.Token,
			Command:       interaction.ApplicationCommandData().Name,
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
		}

		if parsedArguments["User"] == "" {
			destination := "sqsReceiver"

			if parsedArguments["Redditpost"] != "" {
				sqsMessage.Data = parsedArguments["Redditpost"]
				sqsMessage.Type = "redditpost"
				destination = "speakReceiver"
			} else if parsedArguments["Url"] != "" {
				sqsMessage.Type = "url"
				sqsMessage.Data = parsedArguments["Url"]
			} else if parsedArguments["Chat"] != "" {
				sqsMessage.Type = "chat"
				sqsMessage.Data = parsedArguments["Chat"]
				destination = "chatReceiver"
			}
			fmt.Println(destination)

			bot.InteractionResponseEdit(interaction.Interaction, &response)
		} else {
			sqsMessage.Type = "user"
			sqsMessage.Data = parsedArguments["User"]
			err := util.ConfigFile.SendStatsBotRequest(sqsMessage)

			if err != nil {
				fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
				return
			}

		}
	}
}
