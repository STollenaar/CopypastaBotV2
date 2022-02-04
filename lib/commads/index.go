package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

// Bot main reference to the bot
var Bot *discordgo.Session

// commandArgs basic command args
type commandArgs struct {
	Word string `description:"The word to for the command."`
}

// CommandParsed parsed struct for count command
type CommandParsed struct {
	Word          string
	GuildID       string
	UserTarget    string
	ChannelTarget string
}

func Init(bot *discordgo.Session, parser *parsley.Parser) {
	Bot = bot
	PingInit(parser)
	MarkovInit(parser)
	CopyInit(parser)
}
