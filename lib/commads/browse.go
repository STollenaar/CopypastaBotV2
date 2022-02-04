package commands

import (
	"copypastabot/util"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

func BrowseInit(parser *parsley.Parser) {
	parser.NewCommand("browse", "Browse reddit from the comfort of discord", BrowseCommand)
}

// BrowseCommand basic handler
func BrowseCommand(message *discordgo.MessageCreate, args optionalCommandArg) {
	if args.Word != "" {
		args.Word = "all"
	}
	posts := util.DisplayRedditSubreddit(args.Word)

	embed := util.DisplayRedditPost(posts[0].ID, true)[0]
	sendMessage, _ := Bot.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
		Embed: &embed,
	})
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "🔙")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "◀")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "▶")
}
