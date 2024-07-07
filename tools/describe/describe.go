package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	bot   *discordgo.Session
	Guild = flag.String("guild", "", "Provide the GuildID for the describe")
)

type Role struct {
	id, name string
}

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
	if Guild == nil {
		log.Fatal("Required guild flag not provided")
	}
	discordRoles, err := bot.GuildRoles(*Guild)
	if err != nil {
		log.Fatal(err)
	}
	var roles []Role
	for _, role := range discordRoles {
		roles = append(roles, Role{id: role.ID, name: role.Name})
	}

	fmt.Println(roles)

	defer bot.Close()
}
