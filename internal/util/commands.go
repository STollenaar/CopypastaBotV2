package util

import (
	_ "embed"
	"encoding/json"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	//go:embed commands.json
	commandsFile string
	Commands     []*discordgo.ApplicationCommand
)

func init() {
	err := json.Unmarshal([]byte(commandsFile), &Commands)
	if err != nil {
		log.Fatal("Error unmarshalling commands:", err)
	}
}

func ContainsCommand(command string) bool {
	for _, c := range Commands {
		if c.Name == command {
			return true
		}
	}
	return false
}
