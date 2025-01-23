package chat

import (
	_ "embed"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

var (
	//go:embed chatRole.txt
	systemPrompt string
	//go:embed chatRoleSpeak.txt
	systemPromptSpeak string
	//go:embed cavemanRole.txt
	systemCaveman string
)

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})
	if interaction.Data.(discordgo.ApplicationCommandInteractionData).Resolved != nil {
		extractMessageData("message", interaction.Interaction)
	}
	parsedArguments := util.ParseArguments([]string{"message"}, interaction.ApplicationCommandData().Options)

	chatRSP, err := GetChatGPTResponse(interaction.ApplicationCommandData().Name, parsedArguments["Message"], interaction.Member.User.ID)

	if err != nil {
		fmt.Println(fmt.Errorf("error interacting with chatgpt char: %e", err))
		e := "If you see this, and error likely happened. Whoops"
		response := discordgo.WebhookEdit{
			Content: &e,
		}
		bot.InteractionResponseEdit(interaction.Interaction, &response)

		return
	}

	switch interaction.ApplicationCommandData().Name {
	case "respond":
		fallthrough
	case "caveman":
		fallthrough
	case "chat":
		// Getting around the 4096 word limit
		contents := util.BreakContent(chatRSP.Choices[0].Message.Content, 4096)
		var embeds []*discordgo.MessageEmbed
		for _, content := range contents {
			embed := discordgo.MessageEmbed{}
			embed.Description = content
			embeds = append(embeds, &embed)
		}
		response := discordgo.WebhookEdit{
			Embeds: &embeds,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Users: []string{},
				Roles: []string{},
			},
		}
		_, err := bot.InteractionResponseEdit(interaction.Interaction, &response)

		if err != nil {
			fmt.Println(fmt.Errorf("error creating char: %e", err))
		}
	case "respond-vc":
		fallthrough
	case "caveman-vc":
		fallthrough
	case "speak":

	default:
		fmt.Println(fmt.Errorf("unimplemented command: %s", interaction.ApplicationCommandData().Name))
	}

}

func GetChatGPTResponse(promptName, message, userID string) (openai.ChatCompletionResponse, error) {
	var prompt string
	switch promptName {
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

	return util.GetChatGPTResponse(prompt, message, userID)
}

func extractMessageData(optionName string, interaction *discordgo.Interaction) {
	messageData := interaction.Data.(discordgo.ApplicationCommandInteractionData).Resolved.Messages[interaction.Data.(discordgo.ApplicationCommandInteractionData).TargetID].Content
	appData := interaction.ApplicationCommandData()
	appData.Options = append(appData.Options, &discordgo.ApplicationCommandInteractionDataOption{
		Name:  optionName,
		Type:  3,
		Value: messageData,
	})
	interaction.Data = appData
}
