package markovCommand

import (
	"copypastabot/lib/markov"
	"copypastabot/util"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// MarkovUserCommand create a markov chain from an user
func MarkovUserCommand(interaction *discordgo.InteractionCreate, user string) {

	generated, err := getUserMarkov(interaction.GuildID, user)

	if err != nil {
		INVALID_URL_RESPONSE := "Not a valid URL was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &INVALID_URL_RESPONSE,
		})
		return
	}

	Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &generated,
	})
}

// getUserMarkov create a markov chain from an user
func getUserMarkov(guildID, userID string) (string, error) {
	resp, err := http.Get("http://" + statsbotUrl + ":3000/userMessages/" + guildID + "/" + userID)
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
