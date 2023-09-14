package markov

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stollenaar/copypastabot/internal/markov"
	"github.com/stollenaar/copypastabot/internal/util"

	"github.com/bwmarrin/discordgo"
)

// MarkovUserCommand create a markov chain from an user
func MarkovUserCommand(interaction *discordgo.InteractionCreate, user string) {

	generated, err := GetUserMarkov(interaction.GuildID, user)

	if err != nil {
		INVALID_USER_RESPONSE := "Not a valid User was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &INVALID_USER_RESPONSE,
		})
		return
	}

	Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &generated,
	})
}

// GetUserMarkov create a markov chain from an user
func GetUserMarkov(guildID, userID string) (string, error) {
	req := fmt.Sprintf("http://%s:3000/userMessages/%s/%s", util.ConfigFile.STATISTICS_BOT, guildID, userID)
	fmt.Printf("Fetching with request: %s\n", req)
	resp, err := http.Get(req)

	if err != nil {
		log.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var messageObjects []util.MessageObject
	json.Unmarshal(body, &messageObjects)

	messages := mapToContent(&messageObjects)
	messages = filterNonTexts(messages)

	textSeed := strings.Join(messages, " ")
	markov := markov.New()

	generated := markov.ReadText(textSeed)
	return generated, nil
}

func mapToContent(messages *[]util.MessageObject) (result []string) {
	for _, message := range *messages {
		if len(message.Content) == 0 {
			continue
		}
		lastWord := message.Content[len(message.Content)-1]
		if !util.IsTerminalWord(lastWord) {
			lastWord += "."
			message.Content[len(message.Content)-1] = lastWord
		}
		result = append(result, message.Content...)
	}
	return result
}

func filterNonTexts(messages []string) (result []string) {
	for _, message := range messages {
		if len(reTarget.FindAllString(message, -1)) != 3 {
			result = append(result, message)
		}
	}
	return result
}
