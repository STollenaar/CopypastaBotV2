package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/nint8835/parsley"

	"copypastabot/lib/pastaCommand"
	"copypastabot/util"
	"copypatebot/lib/pingCommand"
)

var (
	bot      *discordgo.Session
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "browse",
			Description: "Browse reddit from the confort of discord",
		},
		{
			Name:        "ping",
			Description: "Pong",
		},
		{
			Name:        "pasta",
			Description: "Post a reddit post with an url or post id",
			Options: []*discordgo.ApplicationCommandOption{
				&discordgo.ApplicationCommandOption{
					Name: "URL",
					Type: discordgo.ApplicationCommandOptionString,
				},
				&discordgo.ApplicationCommandOption{
					Name: "PostID",
					Type: discordgo.ApplicationCommandOptionString,
				},
			},
		},
		{
			Name:        "markov",
			Description: "Imitate someone or from a reddit post with some weird results",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping":  pingCommand.Command,
		"pasta": pastaCommand.Command,
	}
)

func init() {
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

	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessageTyping |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuilds)

	parser := parsley.New("pasta!")
	parser.RegisterHandler(bot)

	// lib.Init(bot, parser)
	if err != nil {
		fmt.Println("Error loading command ", err)
		return
	}
}

func main() {
	err := bot.Open()
	if err != nil {
		fmt.Println("Error starting bot ", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	bot.Close()
}
