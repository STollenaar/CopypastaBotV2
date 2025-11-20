package pasta

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	PastaCmd = PastaCommand{
		Name:        "pasta",
		Description: "Post a reddit post with an url or post id",
	}
)

type PastaCommand struct {
	Name        string
	Description string
}

func (p PastaCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	var redditPostID string
	if argUrl, ok := sub.Options["url"]; ok {
		if uri, err := url.ParseRequestURI(argUrl.String()); err == nil {
			path := strings.Split(uri.Path, "/")
			postID := findIndex(path, "comments")
			redditPostID = path[postID+1]
		}
	}
	if postid, ok := sub.Options["postid"]; ok {
		redditPostID = postid.String()
	}
	embeds := util.DisplayRedditPost(redditPostID, false)

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}
}

func (p PastaCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "url",
			Description: "URL of the reddit post",
			Required:    false,
		},
		discord.ApplicationCommandOptionString{
			Name:        "postid",
			Description: "The Reddit postID of the post",
			Required:    false,
		},
	}
}

func findIndex(array []string, param string) int {
	for k, v := range array {
		if v == param {
			return k
		}
	}
	return -1
}

func breakContent(content string) (result []string) {
	words := strings.Split(content, " ")

	var tmp string
	for i, word := range words {
		if i == 0 {
			tmp = word
		} else if len(tmp)+len(word) < 4096 {
			tmp += " " + word
		} else {
			result = append(result, tmp)
			tmp = word
		}
	}
	result = append(result, tmp)
	return result
}

func arrayContainsSub(array []string, param string) bool {
	for _, v := range array {
		if strings.Contains(param, v) {
			return true
		}
	}
	return false
}
