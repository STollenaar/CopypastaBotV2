package markovCommand

import (
	"copypastabot/lib/markov"
	URL "net/url"

	"github.com/bwmarrin/discordgo"
)

// MarkovURLCommand create a markov chain from an URL
func MarkovURLCommand(interaction *discordgo.InteractionCreate, url string) {
	u, err := URL.ParseRequestURI(url)

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
