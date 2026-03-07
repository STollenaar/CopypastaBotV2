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

	switch *sub.SubCommandName {
	case "url":
		if uri, err := url.ParseRequestURI(sub.Options["url"].String()); err == nil {
			path := strings.Split(uri.Path, "/")
			postID := findIndex(path, "comments")
			redditPostID = path[postID+1]
		}
	case "redditpost":
		redditPostID = sub.Options["postid"].String()
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
		discord.ApplicationCommandOptionSubCommand{
			Name:        "url",
			Description: "Run the speech from the text of a site",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "url",
					Description: "URL of the page to make a markov chain from",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "redditpost",
			Description: "Speech from a reddit post",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "postid",
					Description: "Reddit post ID",
					Required:    true,
				},
			},
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
