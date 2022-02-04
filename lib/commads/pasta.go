package commands

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var redditClient *reddit.Client

var images []string
var videos []string

func CopyInit(parser *parsley.Parser) {
	parser.NewCommand("copy", "Copypasta from the provided reddit url or threadID", CopyCommand)
	redditClient, _ = reddit.NewReadonlyClient()
	images = []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".gif"}
	videos = []string{"youtube", "gfycat", "youtu"}
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
	postCommnents, _, _ := redditClient.Post.Get(context.TODO(), args.Word)
	body := postCommnents.Post.Body

	if postCommnents.Post.Body == "" {
		body = postCommnents.Post.URL
	}

	// Getting around the 4096 word limit
	contents := breakContent(body)

	for i, content := range contents {
		embed := discordgo.MessageEmbed{}

		// Making sure if it's not a text content that the embed is set correctly
		if uri, err := url.ParseRequestURI(content); err == nil {
			if arrayContainsSub(images, uri.RequestURI()) {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: content,
				}
			} else if arrayContainsSub(videos, uri.RequestURI()) {
				embed.Video = &discordgo.MessageEmbedVideo{
					URL: content,
				}
			} else {
				embed.Description = content
			}
		} else {
			embed.Description = content
		}

		if i == 0 {
			embed.Title = postCommnents.Post.Title
			embed.Author = &discordgo.MessageEmbedAuthor{
				Name: postCommnents.Post.Author,
				URL:  ("https://www.reddit.com/u/" + postCommnents.Post.Author),
			}
		}

		// Only adding the footer if this is the last entry
		if i == len(contents)-1 {
			embed.Fields = []*discordgo.MessageEmbedField{
				{
					Name:   "Reddit Thread:",
					Value:  ("https://www.reddit.com" + postCommnents.Post.Permalink),
					Inline: true,
				},
			}
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: ("PostID: " + postCommnents.Post.ID),
			}
		}
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
