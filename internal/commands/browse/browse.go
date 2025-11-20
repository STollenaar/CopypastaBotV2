package browse

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	BrowseCmd = BrowseCommand{
		Name:        "browse",
		Description: "Browse reddit from the comfort of discord",
	}
)

type BrowseCommand struct {
	Name        string
	Description string
}

type browserTracker struct {
	UserId    string `json:"userId"`
	SubReddit string `json:"subReddit"`
	Page      int    `json:"page"`
	Action    string `json:"action"`
}

func (b *browserTracker) Marshal() string {
	return fmt.Sprintf("%s|%d|%s|%s", b.SubReddit, b.Page, b.Action, b.UserId)
}

func (b *browserTracker) Unmarshal(data []byte) error {
	d := strings.Split(string(data), "|")
	if len(d) != 4 {
		return errors.New("unknown data format")
	}
	page, _ := strconv.Atoi(d[1])
	b.SubReddit = d[0]
	b.Page = page
	b.Action = d[2]
	b.UserId = d[3]
	return nil
}

func (b BrowseCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	var browser browserTracker

	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	browser = browserTracker{
		SubReddit: sub.Options["subreddit"].String(),
		Page:      0,
		UserId:    event.User().ID.String(),
	}

	posts := util.DisplayRedditSubreddit(browser.SubReddit)
	embed := util.DisplayRedditPost(posts[browser.Page].ID, true)[0]

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds:     &[]discord.Embed{embed},
		Components: getActionRow(browser),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}
}

func (b BrowseCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "subreddit",
			Description: "subreddit",
			Required:    true,
		},
	}
}

func (b BrowseCommand) ComponentHandler(event *events.ComponentInteractionCreate) {
	if event.Message.Interaction.User.ID != event.Member().User.ID {
		return
	}

	err := event.DeferUpdateMessage()

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}
	var browser browserTracker

	err = browser.Unmarshal([]byte(event.Data.CustomID()))
	if err != nil {
		fmt.Printf("Error unmarshalling browser data: %v\n", err)
		return
	}

	posts := util.DisplayRedditSubreddit(browser.SubReddit)
	embed := util.DisplayRedditPost(posts[browser.Page].ID, true)[0]

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds:     &[]discord.Embed{embed},
		Components: getActionRow(browser),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}
}

func getActionRow(browser browserTracker) *[]discord.LayoutComponent {
	prevBrowser := browserTracker{
		UserId:    browser.UserId,
		SubReddit: browser.SubReddit,
		Page:      browser.Page - 1,
		Action:    "previous",
	}

	nextBrowser := browserTracker{
		UserId:    browser.UserId,
		SubReddit: browser.SubReddit,
		Page:      browser.Page + 1,
		Action:    "next",
	}

	if prevBrowser.Page < 0 {
		prevBrowser.Page = 0
	}

	bdataPrev := prevBrowser.Marshal()
	bdataNext := nextBrowser.Marshal()

	return &[]discord.LayoutComponent{
		discord.ActionRowComponent{
			Components: []discord.InteractiveComponent{
				discord.ButtonComponent{
					CustomID: bdataPrev,
					Label:    "Previous",
				},
				discord.ButtonComponent{
					CustomID: bdataNext,
					Label:    "Next",
				},
			},
		},
	}
}
