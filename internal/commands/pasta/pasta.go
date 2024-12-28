package pasta

import (
	"net/url"
	"strings"

	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})

	var response discordgo.WebhookEdit

	parsedArguments := util.ParseArguments([]string{"url", "postid"}, interaction.ApplicationCommandData().Options)

	if parsedArguments["Url"] != "" {
		if uri, err := url.ParseRequestURI(parsedArguments["Url"]); err == nil {
			path := strings.Split(uri.Path, "/")
			postID := findIndex(path, "comments")
			if postID != -1 {
				parsedArguments["Postid"] = path[postID+1]
			}
			parsedArguments["Postid"] = path[postID+1]
		}
	}
	if parsedArguments["Postid"] != "" {
		embeds := util.DisplayRedditPost(parsedArguments["Postid"], false)
		response.Embeds = &embeds
	}
	bot.InteractionResponseEdit(interaction.Interaction, &response)
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
