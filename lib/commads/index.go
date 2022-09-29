package commands

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

type DiscordBot struct {
	*discordgo.Session
}

// Bot main reference to the bot
var Bot DiscordBot

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
	Bot = DiscordBot{bot}

	if os.Getenv("STATSBOT_URL") != "" {
		statsbotUrl = os.Getenv("STATSBOT_URL")
	} else {
		statsbotUrl = "localhost"
	}

	MarkovInit(parser)
	// SpeakInit(parser)
}
