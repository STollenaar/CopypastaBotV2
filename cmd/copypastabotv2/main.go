package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/stollenaar/copypastabotv2/internal/commands/browse"
	"github.com/stollenaar/copypastabotv2/internal/commands/chat"
	"github.com/stollenaar/copypastabotv2/internal/commands/help"
	"github.com/stollenaar/copypastabotv2/internal/commands/markov"
	"github.com/stollenaar/copypastabotv2/internal/commands/pasta"
	"github.com/stollenaar/copypastabotv2/internal/commands/speak"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

var (
	bot *discordgo.Session

	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"browse":     browse.Handler,
		"chat":       chat.Handler,
		"caveman":    chat.Handler,
		"bayman":     chat.Handler,
		"insult":     chat.Handler,
		"respond":    chat.Handler,
		"caveman-vc": chat.Handler,
		"respond-vc": chat.Handler,
		"help":       help.Handler,
		"markov":     markov.Handler,
		"pasta":      pasta.Handler,
		"ping":       PingCommand,
		"speak":      speak.Handler,
	}
)

func init() {
	flag.Parse()

	token, err := util.ConfigFile.GetDiscordToken()
	if err != nil {
		log.Fatal(err)
	}
	bot, _ = discordgo.New("Bot " + token)

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if h, ok := commandHandlers[i.Interaction.Message.Interaction.Name]; ok {
				h(s, i)
			}
		default:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})
}

func main() {
	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentGuildVoiceStates)

	err := bot.Open()
	if err != nil {
		log.Fatal("Error starting bot ", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(util.Commands))
	commands, err := bot.ApplicationCommands(bot.State.User.ID, *GuildID)
	if err != nil {
		log.Printf("Error fetching registered commands: %e", err)
	}
	for i, v := range util.Commands {
		if cmd := containsCommand(v, commands); cmd != nil && optionsEqual(v, cmd) {
			registeredCommands[i] = cmd
			continue
		}
		var cmd *discordgo.ApplicationCommand
		var err error
		
		if v.GuildID != "" {
			cmd, err = bot.ApplicationCommandCreate(bot.State.User.ID, v.GuildID, v)
		} else {
			cmd, err = bot.ApplicationCommandCreate(bot.State.User.ID, *GuildID, v)
		}

		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Bot up and running")
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if *RemoveCommands {
		log.Println("Removing commands...")
		// We need to fetch the commands, since deleting requires the command ID.
		// We are doing this from the returned commands on line 375, because using
		// this will delete all the commands, which might not be desirable, so we
		// are deleting only the commands that we added.
		// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		// if err != nil {
		// 	log.Fatalf("Could not fetch registered commands: %v", err)
		// }

		for _, v := range registeredCommands {
			err := bot.ApplicationCommandDelete(bot.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}

// PingCommand sends back the pong
func PingCommand(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong",
		},
	})
}

func containsCommand(cmd *discordgo.ApplicationCommand, commands []*discordgo.ApplicationCommand) *discordgo.ApplicationCommand {
	for _, c := range commands {
		if cmd.Name == c.Name {
			return c
		}
	}
	return nil
}

// optionsEqual compares the Options slices of two ApplicationCommands.
func optionsEqual(cmd, registered *discordgo.ApplicationCommand) bool {
	if len(cmd.Options) != len(registered.Options) {
		return false
	}
	for i := range cmd.Options {
		if !optionEqual(cmd.Options[i], registered.Options[i]) {
			return false
		}
	}
	return true
}

// optionEqual compares two ApplicationCommandOption objects recursively.
func optionEqual(a, b *discordgo.ApplicationCommandOption) bool {
	if a.Name != b.Name ||
		a.Description != b.Description ||
		a.Type != b.Type ||
		a.Required != b.Required {
		return false
	}

	// Compare choices if available.
	if len(a.Choices) != len(b.Choices) {
		return false
	}
	for i := range a.Choices {
		if a.Choices[i].Name != b.Choices[i].Name ||
			a.Choices[i].Value != b.Choices[i].Value {
			return false
		}
	}

	// Compare sub-options recursively.
	if len(a.Options) != len(b.Options) {
		return false
	}
	for i := range a.Options {
		if !optionEqual(a.Options[i], b.Options[i]) {
			return false
		}
	}
	return true
}
