package commands

import (
	"copypastabot/util"
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
)

func CopyInit(parser *parsley.Parser) {
	parser.NewCommand("copy", "Copypasta from the provided reddit url or threadID", CopyCommand)
}

// CopyCommand basic multi handler of the copypasta command
func CopyCommand(message *discordgo.MessageCreate, args commandArgs) {
	Bot.ChannelTyping(message.ChannelID)

	if uri, err := url.ParseRequestURI(args.Word); err == nil {
		path := strings.Split(uri.Path, "/")
		postID := findIndex(path, "comments")
		if postID == -1 {
			Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Something went wrong"))
			return
		}
		args.Word = path[postID+1]
	}
	for _, embed := range util.DisplayRedditPost(args.Word, false) {
		Bot.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
			Embed: &embed,
		})
	}
}

func findIndex(array []string, param string) int {
	for k, v := range array {
		if v == param {
			return k
		}
	}
	return -1
}

func breakContent(content string) (result []string) {
	words := strings.Split(content, " ")

	var tmp string
	for i, word := range words {
		if i == 0 {
			tmp = word
		} else if len(tmp)+len(word) < 4096 {
			tmp += " " + word
		} else {
			result = append(result, tmp)
			tmp = word
		}
	}
	result = append(result, tmp)
	return result
}

func arrayContainsSub(array []string, param string) bool {
	for _, v := range array {
		if strings.Contains(param, v) {
			return true
		}
	}
	return false
}
