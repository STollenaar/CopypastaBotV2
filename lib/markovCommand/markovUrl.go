package markovCommand

import (
	"copypastabot/lib/markov"
	"copypastabot/util"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

var (
	Bot *discordgo.Session
)

// Command create a markov chain from an URL
func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if Bot == nil {
		Bot = bot
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading Markov...",
		},
	})

	parsedArguments := util.ParseArguments([]string{"url"}, interaction)
	u, err := url.ParseRequestURI(parsedArguments["Url"])

	if err != nil {
		INVALID_URL_RESPONSE := "Not a valid URL was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &INVALID_URL_RESPONSE,
		})
		return
	}

	markov := markov.New()

	generated := markov.ReadURL(u.String())
	Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &generated,
	})
}
