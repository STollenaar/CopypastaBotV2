package markov

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/util"
	pkgMarkov "github.com/stollenaar/copypastabotv2/pkg/markov"
)

var (
	MarkovCmd = MarkovCommand{
		Name:        "markov",
		Description: "Imitate someone or from a reddit post with some weird results",
	}
)

type MarkovCommand struct {
	Name        string
	Description string
}

func (m MarkovCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	var markovData string

	switch *sub.SubCommandName {
	case "url":
		markovData = HandleURL(sub.Options["url"].String())
	case "user":
		sqsMessage := util.Object{
			GuildID:       event.GuildID().String(),
			ApplicationID: event.ApplicationID().String(),
			Command:       "markov",
			Type:          "user",
			Data:          sub.Options["user"].Snowflake().String(),
			ChannelID:     event.Channel().ID().String(),
			Token:         event.Token(),
		}

		resp, err := util.ConfigFile.SendStatsBotRequest(sqsMessage)
		if err != nil {
			slog.Error("Error sending statsbot request", slog.Any("err", err))
			return
		}
		markovData = HandleUser(resp.Data)
	}

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Content: &markovData,
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}
}

func (m MarkovCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
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
			Name:        "user",
			Description: "Run the speech from one of the user messages",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "User to create a markov chain of",
					Required:    true,
				},
			},
		},
	}
}

func HandleURL(input string) string {
	data, err := pkgMarkov.GetMarkovURL(input)
	if err != nil {
		slog.Error("Error getting markov from URL", slog.Any("err", err))
		return err.Error()
	}
	return data
}

func HandleUser(input string) string {
	data, err := pkgMarkov.GetUserMarkov(input)
	if err != nil {
		slog.Error("Error getting user markov", slog.Any("err", err))
		return err.Error()
	}
	return data
}
