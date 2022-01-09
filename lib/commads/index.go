package commands

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

// Bot main reference to the bot
var Bot *discordgo.Session

// commandArgs basic command args
type commandArgs struct {
	Word        string `description:"The word to for the command."`
	Target      string `default:"" description:"The first to filter for."`
	TargetOther string `default:"" description:"The second to filter for."`
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
}

// parseArguments parses the arguments from the command into an unified struct
func parseArguments(message *discordgo.Message, args commandArgs) (parsedArguments CommandParsed) {
	reTarget := regexp.MustCompile("[\\<>@#&!]")

	parsedArguments = CommandParsed{Word: args.Word, GuildID: message.GuildID}

	if args.Target != "" {
		if strings.Contains(args.Target, "@") {
			parsedArguments.UserTarget = reTarget.ReplaceAllString(args.Target, "")
		} else if strings.Contains(args.Target, "#") {
			parsedArguments.ChannelTarget = reTarget.ReplaceAllString(args.Target, "")
		}
	} else {
		parsedArguments.UserTarget = message.Author.ID
	}

	if args.TargetOther != "" {
		if strings.Contains(args.TargetOther, "@") {
			parsedArguments.UserTarget = reTarget.ReplaceAllString(args.TargetOther, "")
		} else if strings.Contains(args.TargetOther, "#") {
			parsedArguments.ChannelTarget = reTarget.ReplaceAllString(args.TargetOther, "")
		}
	}

	return parsedArguments
}
