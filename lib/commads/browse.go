package commands

import (
	"copypastabot/util"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type browserTracker struct {
	messageID string
	channelID string
	subReddit string
	postIDs   []string
	page      int
}

var activeBrowsers map[string]*browserTracker

func BrowseInit(parser *parsley.Parser) {
	parser.NewCommand("browse", "Browse reddit from the comfort of discord", BrowseCommand)
	activeBrowsers = make(map[string]*browserTracker)
	Bot.AddHandler(BrowseHandler)
}

// BrowseCommand basic handler
func BrowseCommand(message *discordgo.MessageCreate, args optionalCommandArg) {
	Bot.ChannelTyping(message.ChannelID)
	if args.Word != "" {
		args.Word = "all"
	}

	if activeBrowsers[message.Author.ID] != nil {
		userBrowserSession := activeBrowsers[message.Author.ID]
		Bot.ChannelMessageDelete(userBrowserSession.channelID, userBrowserSession.messageID)
	}

	posts := util.DisplayRedditSubreddit(args.Word)

	embed := util.DisplayRedditPost(posts[0].ID, true)[0]

	sendMessage, _ := Bot.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
		Embed: &embed,
	})

	activeBrowsers[message.Author.ID] = &browserTracker{
		messageID: sendMessage.ID,
		channelID: sendMessage.ChannelID,
		subReddit: args.Word,
		page:      0,
		postIDs:   mapToID(posts),
	}

	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "🔙")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "◀")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "▶")

	time.AfterFunc(1*time.Hour, func() {
		Bot.ChannelMessageDelete(message.ChannelID, message.ID)
		delete(activeBrowsers, message.Author.ID)
	})
}

func BrowseHandler(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	// Fetch some extra information about the message associated to the reaction
	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)

	// Ignore reactions on messages that have an error or that have not been sent by the bot
	if err != nil || message == nil || message.Author.ID != session.State.User.ID {
		return
	}

	// Ignore messages that are not embeds with a command in the footer
	if len(message.Embeds) != 1 || message.Embeds[0].Footer == nil || message.Embeds[0].Footer.Text == "" {
		return
	}

	// Ignore reactions that haven't been set by the bot
	if !isBotReaction(session, message.Reactions, &reaction.Emoji) {
		return
	}

	user, err := session.User(reaction.UserID)
	// Ignore when sender is invalid or is a bot
	if err != nil || user == nil || user.Bot {
		return
	}

	// Check if this message still has the reactions on it after a while
	if messageTime := message.Timestamp; time.Now().After(messageTime.Add(1 * time.Hour)) {
		Bot.ChannelMessageDelete(message.ChannelID, message.ID)
		return
	}

	// Check if this message is being tracked for the user
	if activeBrowsers[reaction.UserID] == nil || activeBrowsers[reaction.UserID].messageID != message.ID {
		// Checking if this message is an old one left over from a restart
		go func(messageID, channelID string) {
			checkOrphanedBrowsers(messageID, channelID)
		}(message.ID, message.ChannelID)
		return
	}

	doEmbedHandler(message, reaction.MessageReaction)
}

// Cleaning up old embeds when the bot had to restart
func checkOrphanedBrowsers(messageID, channelID string) {
	for _, v := range activeBrowsers {
		if v.messageID == messageID {
			return
		}
	}
	Bot.ChannelMessageDelete(channelID, messageID)
}

func mapToID(posts []*reddit.Post) (IDs []string) {
	for _, post := range posts {
		IDs = append(IDs, post.ID)
	}
	return IDs
}

func isBotReaction(session *discordgo.Session, reactions []*discordgo.MessageReactions, emoji *discordgo.Emoji) bool {
	for _, reaction := range reactions {
		if reaction.Emoji.Name == emoji.Name && reaction.Me {
			return true
		}
	}
	return false
}

// handling the post change
func doEmbedHandler(message *discordgo.Message, reaction *discordgo.MessageReaction) {
	Bot.MessageReactionAdd(message.ChannelID, message.ID, "🔄")

	userBrowserSession := activeBrowsers[reaction.UserID]

	switch reaction.Emoji.Name {
	case "🔙":
		userBrowserSession.page = 0
	case "◀":
		userBrowserSession.page -= 0
		if userBrowserSession.page < 0 {
			userBrowserSession.page = 0
		}
	case "▶":
		userBrowserSession.page += 1
	}

	embed := util.DisplayRedditPost(userBrowserSession.postIDs[userBrowserSession.page], true)[0]

	Bot.ChannelMessageEditEmbed(userBrowserSession.channelID, userBrowserSession.messageID, &embed)
	Bot.MessageReactionRemove(userBrowserSession.channelID, userBrowserSession.messageID, reaction.Emoji.Name, reaction.UserID)
	Bot.MessageReactionRemove(userBrowserSession.channelID, userBrowserSession.messageID, "🔄", Bot.State.User.ID)
}
