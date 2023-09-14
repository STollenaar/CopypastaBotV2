package pasta

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/stollenaar/copypastabot/internal/util"

	"github.com/bwmarrin/discordgo"
)

func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.ChannelTyping(interaction.ChannelID)

	parsedArguments := util.ParseArguments([]string{"URL", "PostID"}, interaction)

	if parsedArguments["URL"] != "" {
		if uri, err := url.ParseRequestURI(parsedArguments["URL"]); err == nil {
			path := strings.Split(uri.Path, "/")
			postID := findIndex(path, "comments")
			if postID == -1 {
				bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintln("Something went wrong"),
					},
				})
				return
			}
			parsedArguments["PostID"] = path[postID+1]
		}
	}
	if parsedArguments["PostID"] != "" {
		embeds := util.DisplayRedditPost(parsedArguments["PostID"], false)
		bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
	} else {
		// Error
		bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintln("Something went wrong"),
			},
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
