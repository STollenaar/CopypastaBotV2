package markov

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	statsbotUrl string
	reTarget    *regexp.Regexp
)

func init() {
	reTarget = regexp.MustCompile(`[\<>@#&!]`)
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
	markov := New()

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
