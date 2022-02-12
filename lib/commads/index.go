package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

// Bot main reference to the bot
var Bot *discordgo.Session

// optionalCommandArg other basic command arg
type optionalCommandArg struct {
	Word string `default:"" description:"The word to for the command."`
}

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
	BrowseInit(parser)
}
