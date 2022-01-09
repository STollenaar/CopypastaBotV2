package commands

import (
	"copypastabot/lib/markov"
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

// MarkovURLCommand create a markov chain from an URL
func MarkovURLCommand(message *discordgo.MessageCreate, args commandArgs) {

	u, err := url.ParseRequestURI(args.Word)
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Not a valid URL was provided"))
		return
	}

	markov := markov.New()

	Bot.ChannelMessageSend(message.ChannelID, markov.ReadURL(u.String()))
}
