package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

func MarkovInit(parser *parsley.Parser) {
	parser.NewCommand("markov", "Markov pasta", MarkovCommand)
}

// MarkovCommand create a markov chain from an URL
func MarkovCommand(message *discordgo.MessageCreate, args commandArgs) {
	Bot.ChannelTyping(message.ChannelID)

	if _, err := url.ParseRequestURI(args.Word); err == nil {
		MarkovURLCommand(message, args, true)
	} else if strings.Contains(args.Word, "@") {
		MarkovUserCommand(message, args, true)
	} else {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Not a argument was provided"))
	}

}
