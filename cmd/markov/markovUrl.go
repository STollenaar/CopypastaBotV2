package main

import (
	URL "net/url"

	"github.com/stollenaar/copypastabotv2/internal/markov"

	"github.com/bwmarrin/discordgo"
)

// MarkovURLCommand create a markov chain from an URL
func MarkovURLCommand(interaction *discordgo.InteractionCreate, url string) {
	generated, err := GetMarkovURL(url)
	if err != nil {
		INVALID_URL_RESPONSE := "Not a valid URL was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &INVALID_URL_RESPONSE,
		})
	} else {
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &generated,
		})
	}
}

// GetMarkovURL creating a markov chain from a provided url
func GetMarkovURL(url string) (string, error) {
	u, err := URL.ParseRequestURI(url)

	if err != nil {
		return "", err
	}
	markov := markov.New()

	generated := markov.ReadURL(u.String())
	return generated, nil
}
