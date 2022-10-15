package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/nint8835/parsley"

	"copypastabot/lib/browseCommand"
	"copypastabot/lib/markovCommand"
	"copypastabot/lib/pastaCommand"
	"copypastabot/lib/pingCommand"
	"copypastabot/util"
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
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Description: "User to create a markov chain of",
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
	}
)

func init() {
	flag.Parse()

	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading environment variables")
		return
	}

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

}

func init() {

	parser := parsley.New("pasta!")
	parser.RegisterHandler(bot)

	// lib.Init(bot, parser)
}

func main() {
	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessageTyping |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuilds)

	err := bot.Open()
	if err != nil {
		fmt.Println("Error starting bot ", err)
		return
	}
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
