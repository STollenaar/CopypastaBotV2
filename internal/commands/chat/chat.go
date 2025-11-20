package chat

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	//go:embed chatRole.txt
	systemPrompt string
	//go:embed chatRoleSpeak.txt
	systemPromptSpeak string
	//go:embed cavemanRole.txt
	systemCaveman string
	//go:embed insultRole.txt
	systemInsult string

	ChatCmd = ChatCommand{
		Name:        "chat",
		Description: "Chat with the bot",
		MessageCommands: []string{"caveman", "insult", "respond"},
	}
)

type ChatResponse struct {
	Response string `json:"response"`
}

type ChatCommand struct {
	Name            string
	Description     string
	MessageCommands []string
}

func (c ChatCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()
	chatRSP, err := GetChatGPTResponse(event.Data.CommandName(), sub.Options["message"].String(), event.User().ID.String())

	if err != nil {
		fmt.Println(fmt.Errorf("error interacting with chatgpt char: %e", err))
		e := "If you see this, and error likely happened. Whoops"

		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &e,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}
		return
	}

	// Getting around the 4096 word limit
	contents := util.BreakContent(chatRSP.Response, 4096)
	var embeds []discord.Embed
	for _, content := range contents {
		embed := discord.Embed{}
		embed.Description = content
		embeds = append(embeds, embed)
	}

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Embeds: &embeds,
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}

}

func (c ChatCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "message",
			Description: "message for the bot",
			Required:    true,
		},
	}
}

func (c ChatCommand) MessageCommandHandler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	chatRSP, err := GetChatGPTResponse(event.Data.CommandName(), event.MessageCommandInteractionData().TargetMessage().Content, event.User().ID.String())

	if err != nil {
		fmt.Println(fmt.Errorf("error interacting with chatgpt char: %e", err))
		e := "If you see this, and error likely happened. Whoops"

		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &e,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}
		return
	}

	switch event.Data.CommandName() {
	case "bayman":
		fallthrough
	case "respond":
		fallthrough
	case "caveman":
		fallthrough
	case "chat":
		fallthrough
	case "insult":
		// Getting around the 4096 word limit
		contents := util.BreakContent(chatRSP.Response, 4096)
		var embeds []discord.Embed
		for _, content := range contents {
			embed := discord.Embed{}
			embed.Description = content
			embeds = append(embeds, embed)
		}

		_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Embeds: &embeds,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}

	case "respond-vc":
		fallthrough
	case "caveman-vc":
		fallthrough
	case "speak":

	default:
		fmt.Println(fmt.Errorf("unimplemented command: %s", event.Data.CommandName()))
	}
}

func GetChatGPTResponse(promptName, message, userID string) (out ChatResponse, err error) {
	var prompt string
	switch promptName {
	case "insult":
		prompt = systemInsult
	case "caveman":
		fallthrough
	case "caveman-vc":
		prompt = systemCaveman
	case "respond":
		fallthrough
	case "chat":
		prompt = systemPrompt
	case "respond-vc":
		fallthrough
	case "speak":
		prompt = systemPromptSpeak
	}
	prompt = fmt.Sprintf(prompt, userID)
	resp, err := util.CreateOllamaGeneration(util.OllamaGenerateRequest{
		Model:  util.ConfigFile.OLLAMA_MODEL,
		Prompt: prompt,
		Stream: false,
		Format: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"response": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{
				"response",
			},
		},
	})

	if err != nil {
		return ChatResponse{}, nil
	}

	err = json.Unmarshal([]byte(resp.Response), &out)
	return

}
