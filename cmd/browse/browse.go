package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type browserTracker struct {
	messageID string
	channelID string
	subReddit string
	postIDs   []string
	page      int
}

var (
	activeBrowsers map[string]*browserTracker
	Bot            *discordgo.Session
)

func init() {
	activeBrowsers = make(map[string]*browserTracker)
}

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := discordgo.InteractionResponseData{
		Content: "Loading...",
	}
	data, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}

func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if Bot == nil {
		Bot = bot
		Bot.AddHandler(BrowseHandler)
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading Posts...",
		},
	})

	parsedArguments := util.ParseArguments([]string{"name"}, interaction)
	if parsedArguments["Name"] == "" {
		parsedArguments["Name"] = "all"
	}
	userID := interaction.Member.User.ID
	if userBrowserSession := activeBrowsers[userID]; userBrowserSession != nil {
		Bot.ChannelMessageDelete(userBrowserSession.channelID, userBrowserSession.messageID)
	}

	posts := util.DisplayRedditSubreddit(parsedArguments["Name"])
	embed := util.DisplayRedditPost(posts[0].ID, true)[0]

	sendMessage, _ := Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	activeBrowsers[userID] = &browserTracker{
		messageID: sendMessage.ID,
		channelID: sendMessage.ChannelID,
		subReddit: parsedArguments["Name"],
		page:      0,
		postIDs:   mapToID(posts),
	}

	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "🔙")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "◀")
	Bot.MessageReactionAdd(sendMessage.ChannelID, sendMessage.ID, "▶")

	time.AfterFunc(1*time.Hour, func() {
		Bot.ChannelMessageDelete(sendMessage.ChannelID, sendMessage.ID)
		delete(activeBrowsers, userID)
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

	Bot.ChannelMessageEditEmbed(userBrowserSession.channelID, userBrowserSession.messageID, embed)
	Bot.MessageReactionRemove(userBrowserSession.channelID, userBrowserSession.messageID, reaction.Emoji.Name, reaction.UserID)
	Bot.MessageReactionRemove(userBrowserSession.channelID, userBrowserSession.messageID, "🔄", Bot.State.User.ID)
}
