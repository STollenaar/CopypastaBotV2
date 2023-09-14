package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/stollenaar/copypastabotv2/lib/browseCommand"
	"github.com/stollenaar/copypastabotv2/lib/markovCommand"
	"github.com/stollenaar/copypastabotv2/lib/pastaCommand"
	"github.com/stollenaar/copypastabotv2/lib/pingCommand"
	"github.com/stollenaar/copypastabotv2/lib/speakCommand"
	"github.com/stollenaar/copypastabotv2/util"
)

var (
	bot *discordgo.Session

	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "browse",
			Description: "Browse reddit from the confort of discord",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Name of the subreddit",
					Required:    false,
				},
			},
		},
		{
			Name:        "ping",
			Description: "Pong",
		},
		{
			Name:        "pasta",
			Description: "Post a reddit post with an url or post id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "url",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "URL of the reddit post",
					Required:    false,
				},
				{
					Name:        "postid",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "The Reddit postID of the post",
					Required:    false,
				},
			},
		},
		{
			Name:        "markov",
			Description: "Imitate someone or from a reddit post with some weird results",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "url",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "URL of the page to make a markov chain from",
					Required:    false,
				},
				{
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "User to create a markov chain of",
					Required:    false,
				},
			},
		},
		{
			Name:        "speak",
			Description: "Imitate someone or from a reddit post with some weird results, and listen to the beauty",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "url",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "URL of the page to make a markov chain from",
					Required:    false,
				},
				{
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "User to create a markov chain of",
					Required:    false,
				},
				{
					Name:        "redditpost",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Reddit post id",
					Required:    false,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"browse": browseCommand.Command,
		"pasta":  pastaCommand.Command,
		"ping":   pingCommand.Command,
		"markov": markovCommand.Command,
		"speak":  speakCommand.Command,
	}
)

func init() {
	flag.Parse()
	var err error
	bot, err = discordgo.New("Bot " + util.ConfigFile.GetDiscordToken())
	if err != nil {
		fmt.Println("Error loading bot ", err)
		return
	}

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	// bot.AddHandler(onEvent)

	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessageTyping |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuilds)
}

func main() {

	err := bot.Open()
	if err != nil {
		fmt.Println("Error starting bot ", err)
		return
	}

	speakCommand.VCInterupt(bot)
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := bot.ApplicationCommandCreate(bot.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc
	bot.Close()
}

func onEvent(session *discordgo.Session, event *discordgo.Event) {
	fmt.Println(event.Type)
}
