package commands

import (
	"reflect"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/commands/browse"
	"github.com/stollenaar/copypastabotv2/internal/commands/chat"
	"github.com/stollenaar/copypastabotv2/internal/commands/markov"
	"github.com/stollenaar/copypastabotv2/internal/commands/pasta"
	"github.com/stollenaar/copypastabotv2/internal/commands/speak"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

type CommandI interface {
	Handler(e *events.ApplicationCommandInteractionCreate)
	CreateCommandArguments() []discord.ApplicationCommandOption
}

var (
	Commands               = []CommandI{browse.BrowseCmd, chat.ChatCmd, markov.MarkovCmd, pasta.PastaCmd, speak.SpeakCmd}
	ApplicationCommands    []discord.ApplicationCommandCreate
	CommandHandlers        = make(map[string]func(e *events.ApplicationCommandInteractionCreate))
	MessageCommandHandlers = make(map[string]func(e *events.ApplicationCommandInteractionCreate))
	ModalSubmitHandlers    = make(map[string]func(e *events.ModalSubmitInteractionCreate))
	ComponentHandlers      = make(map[string]func(e *events.ComponentInteractionCreate))
)

func init() {
	for _, cmd := range Commands {
		ApplicationCommands = append(ApplicationCommands, discord.SlashCommandCreate{
			Name:        reflect.ValueOf(cmd).FieldByName("Name").String(),
			Description: reflect.ValueOf(cmd).FieldByName("Description").String(),
			Options:     cmd.CreateCommandArguments(),
		})
		CommandHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = cmd.Handler

		if _, ok := reflect.TypeOf(cmd).MethodByName("ModalHandler"); ok {
			ModalSubmitHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = func(e *events.ModalSubmitInteractionCreate) {
				reflect.ValueOf(cmd).MethodByName("ModalHandler").Call([]reflect.Value{
					reflect.ValueOf(e),
				})
			}
		}
		if _, ok := reflect.TypeOf(cmd).MethodByName("ComponentHandler"); ok {
			ComponentHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = func(e *events.ComponentInteractionCreate) {
				reflect.ValueOf(cmd).MethodByName("ComponentHandler").Call([]reflect.Value{
					reflect.ValueOf(e),
				})
			}
		}

		if _, ok := reflect.TypeOf(cmd).FieldByName("MessageCommands"); ok {
			messageCmds := reflect.ValueOf(cmd).FieldByName("MessageCommands")
			for i := 0; i < messageCmds.Len(); i++ {
				messageCmd := messageCmds.Index(i).String()
				ApplicationCommands = append(ApplicationCommands, discord.MessageCommandCreate{
					Name: messageCmd,
				})
				MessageCommandHandlers[messageCmd] = func(e *events.ApplicationCommandInteractionCreate) {
					reflect.ValueOf(cmd).MethodByName("MessageCommandHandler").Call([]reflect.Value{
						reflect.ValueOf(e),
					})
				}
			}
		}
	}

	ApplicationCommands = append(ApplicationCommands,
		discord.SlashCommandCreate{
			Name:        "ping",
			Description: "pong",
		},
	)

	CommandHandlers["ping"] = PingCommand
}

// PingCommand sends back the pong
func PingCommand(event *events.ApplicationCommandInteractionCreate) {
	event.CreateMessage(discord.MessageCreate{
		Content: "Pong",
		Flags:   util.ConfigFile.SetEphemeral(),
	})
}
