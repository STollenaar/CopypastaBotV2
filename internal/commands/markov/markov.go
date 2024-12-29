package markov

import (
	"fmt"
	"strings"

	"github.com/stollenaar/copypastabotv2/internal/util"
	pkgMarkov "github.com/stollenaar/copypastabotv2/pkg/markov"

	"github.com/bwmarrin/discordgo"
)

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})

	parsedArguments := util.ParseArguments([]string{"url", "user"}, interaction.ApplicationCommandData().Options)

	var ok bool
	var url, user, markovData string
	if url, ok = parsedArguments["Url"]; ok {
		markovData = HandleURL(url)
	} else if user, ok = parsedArguments["User"]; ok {
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")

		sqsMessage := util.Object{
			GuildID:       interaction.GuildID,
			ApplicationID: interaction.AppID,
			Command:       "markov",
			Type:          "user",
			Data:          user,
			ChannelID:     interaction.ChannelID,
			Token:         interaction.Token,
		}

		resp, err := util.ConfigFile.SendStatsBotRequest(sqsMessage)
		if err != nil {
			fmt.Println(err)
			return
		}
		markovData = HandleUser(resp.Data)
	}
	fmt.Println(url, user)
	response := discordgo.WebhookEdit{
		Content: &markovData,
	}

	bot.InteractionResponseEdit(interaction.Interaction, &response)
}

func HandleURL(input string) string {
	data, err := pkgMarkov.GetMarkovURL(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}

func HandleUser(input string) string {
	data, err := pkgMarkov.GetUserMarkov(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}
