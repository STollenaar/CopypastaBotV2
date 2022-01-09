package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

func PingInit(parser *parsley.Parser) {
	parser.NewCommand("ping", "pong", PingCommand)
}

// PingCommand sends back the pong
func PingCommand(message *discordgo.MessageCreate, args struct{}) {
	Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Pong"))
}
