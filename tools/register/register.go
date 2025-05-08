package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	bot            *discordgo.Session
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
)

func init() {
	flag.Parse()

	token, err := util.ConfigFile.GetDiscordToken()
	if err != nil {
		log.Fatal(err)
	}
	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	err = bot.Open()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer bot.Close()
	if *RemoveCommands {
		cmds, err := bot.ApplicationCommands(bot.State.User.ID, "")
		if err != nil {
			log.Panicf("Cannot retrieve commands: %v", err)
		}
		for _, cmd := range cmds {
			err := bot.ApplicationCommandDelete(bot.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", cmd.Name, err)
			}
		}
		fmt.Println("Successfully removed all commands")
	} else {
		for _, v := range util.Commands {
			_, err := bot.ApplicationCommandCreate(bot.State.User.ID, "", v)
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			}
		}
		fmt.Println("Successfully created all commands")
	}
}
