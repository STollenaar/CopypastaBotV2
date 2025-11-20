package markov

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/util"
	pkgMarkov "github.com/stollenaar/copypastabotv2/pkg/markov"
)

var (
	MarkovCmd = MarkovCommand{
		Name: "markov",
		Description: "Imitate someone or from a reddit post with some weird results",
	}
)

type MarkovCommand struct {
	Name string
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

	if url, ok := sub.Options["url"]; ok {
		markovData = HandleURL(url.String())
	} else if user, ok := sub.Options["user"]; ok {
		sqsMessage := util.Object{
			GuildID:       event.GuildID().String(),
			ApplicationID: event.ApplicationID().String(),
			Command:       "markov",
			Type:          "user",
			Data:          user.Snowflake().String(),
			ChannelID:     event.Channel().ID().String(),
			Token:         event.Token(),
		}

		resp, err := util.ConfigFile.SendStatsBotRequest(sqsMessage)
		if err != nil {
			fmt.Println(err)
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
		discord.ApplicationCommandOptionString{
			Name:        "url",
			Description: "URL of the page to make a markov chain from",
			Required:    false,
		},
		discord.ApplicationCommandOptionUser{
			Name:        "user",
			Description: "User the create a markov chain of",
			Required:    false,
		},
	}
}

func HandleURL(input string) string {
	data, err := pkgMarkov.GetMarkovURL(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}

func HandleUser(input string) string {
	data, err := pkgMarkov.GetUserMarkov(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}
