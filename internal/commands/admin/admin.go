package admin

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/commands/speak"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	AdminCmd = AdminCommand{
		Name:        "admin",
		Description: "Admin command to manage to copypastabot",
	}
)

type AdminCommand struct {
	Name        string
	Description string
}

func (a AdminCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	if event.Member().User.ID.String() != util.ConfigFile.ADMIN_USER_ID {
		event.CreateMessage(discord.MessageCreate{
			Content: "You are not the boss of me",
			Flags:   discord.MessageFlagEphemeral | discord.MessageFlagIsComponentsV2,
		})
		return
	}
	err := event.DeferCreateMessage(true)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}
	sub := event.SlashCommandInteractionData()

	switch *sub.SubCommandGroupName {
	case "speak":
		speakHandler(sub, event)
	}
}

func (a AdminCommand) ComponentHandler(event *events.ComponentInteractionCreate) {}

func (a AdminCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "speak",
			Description: "speak subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "queue",
					Description: "Show the current queue",
				},
				{
					Name:        "flush",
					Description: "empty the queue",
				},
				{
					Name:        "skip",
					Description: "skip the current playing object",
				},
			},
		},
	}
}

func speakHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) {
	guildID := event.GuildID().String()

	switch *args.SubCommandName {
	case "queue":
		items := speak.GetQueueItems(guildID)
		embed := discord.Embed{
			Title: fmt.Sprintf("Speak Queue (%d pending)", len(items)),
		}
		if len(items) == 0 {
			embed.Description = "Queue is empty"
		} else {
			for _, item := range items {
				embed.Fields = append(embed.Fields, discord.EmbedField{
					Name:  fmt.Sprintf("#%d · %s", item.Position, item.Type),
					Value: fmt.Sprintf("<@%s>\n> %s", item.UserID, item.Snippet),
				})
			}
		}
		_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Embeds: &[]discord.Embed{embed},
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}

	case "flush":
		speak.FlushQueue(guildID)
		content := "Queue flushed"
		_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &content,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}

	case "skip":
		speak.SkipCurrent(guildID)
		content := "Skipping current item"
		_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &content,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}
	}
}
