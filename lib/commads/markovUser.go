package commands

import (
	"copypastabot/lib/markov"
	"copypastabot/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var reTarget = regexp.MustCompile("[\\<>@#&!]")
var statsbotUrl string

// MarkovUseCommand create a markov chain from an URL
func MarkovUserCommand(message *discordgo.MessageCreate, args commandArgs, channelSend bool) (string, error) {
	userID := reTarget.ReplaceAllString(args.Word, "")

	resp, err := http.Get("http://" + statsbotUrl + ":3000/userMessages/" + message.GuildID + "/" + userID)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		if channelSend {
			Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Something went wrong"))
		}
		return "", err
	}

	var messageObjects []util.MessageObject
	json.Unmarshal(body, &messageObjects)

	messages := mapToContent(&messageObjects)
	messages = filterNonTexts(messages)

	textSeed := strings.Join(messages, " ")
	markov := markov.New()

	generated := markov.ReadText(textSeed)
	if channelSend {
		Bot.ChannelMessageSend(message.ChannelID, generated)
	}
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
