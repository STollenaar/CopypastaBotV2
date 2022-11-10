package markovCommand

import (
	"copypastabot/util"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	Bot         *discordgo.Session
	statsbotUrl string
	reTarget    *regexp.Regexp
)

func init() {
	reTarget = regexp.MustCompile(`[\<>@#&!]`)
}

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

	parsedArguments := util.ParseArguments([]string{"url", "user"}, interaction)

	if url, ok := parsedArguments["Url"]; ok {
		MarkovURLCommand(interaction, url)
	} else if user, ok := parsedArguments["User"]; ok {
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")
		MarkovUserCommand(interaction, user)
	} else {
		unknownState := "Unknown state entered"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &unknownState,
		})
	}
}
