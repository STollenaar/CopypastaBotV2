package commands

import (
	"copypastabot/lib/markov"
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

// MarkovURLCommand create a markov chain from an URL
func MarkovURLCommand(message *discordgo.MessageCreate, args commandArgs, channelSend bool) (string, error) {

	u, err := url.ParseRequestURI(args.Word)
	if err != nil {
		if channelSend {
			Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Not a valid URL was provided"))
		}
		return "", err
	}

	markov := markov.New()

	generated := markov.ReadURL(u.String())
	if channelSend {
		Bot.ChannelMessageSend(message.ChannelID, generated)
	}
	return generated, nil
}
