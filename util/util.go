package util

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var termRegex string
var images []string
var videos []string
var redditClient *reddit.Client

func IsTerminalWord(word string) bool {
	if termRegex == "" {
		termRegex = os.Getenv("TERMINAL_REGEX")

		// Setting the default regex
		if termRegex == "" {
			termRegex = `(\.|,|:|;|\?|!)$`
		}
	}
	compiled, _ := regexp.MatchString(termRegex, word)
	return compiled
}

func DisplayRedditSubreddit(subreddit string) []*reddit.Post {
	if redditClient == nil {
		redditClient, _ = reddit.NewReadonlyClient()
	}
	posts, _, _ := redditClient.Subreddit.HotPosts(context.TODO(), subreddit, &reddit.ListOptions{})
	return posts
}

func GetRedditPost(redditPostID string) *reddit.PostAndComments {
	if len(images) == 0 {
		images = []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".gif"}
		videos = []string{"youtube", "gfycat", "youtu"}
	}
	if redditClient == nil {
		redditClient, _ = reddit.NewClient(reddit.Credentials{
			// userAgent: process.env.REDDIT_USER_AGENT,
			Username: os.Getenv("REDDIT_USERNAME"),
			Password: os.Getenv("REDDIT_PASSWORD"),
			ID:       os.Getenv("REDDIT_CLIENT_ID"),
			Secret:   os.Getenv("REDDIT_CLIENT_SECRET"),
		})
	}
	postCommnents, _, err := redditClient.Post.Get(context.TODO(), redditPostID)
	if err != nil {
		fmt.Println(err)
	}

	return postCommnents
}

func DisplayRedditPost(redditPostID string, singleEmbed bool) (embeds []discordgo.MessageEmbed) {
	postCommnents := GetRedditPost(redditPostID)

	body := postCommnents.Post.Body

	if postCommnents.Post.Body == "" {
		body = postCommnents.Post.URL
	}

	// Getting around the 4096 word limit
	contents := BreakContent(body, 4096)
	if singleEmbed {
		contents = []string{contents[0]}
	}

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
		embeds = append(embeds, embed)
	}
	return embeds
}

func BreakContent(content string, maxLength int) (result []string) {
	words := strings.Split(content, " ")

	var tmp string
	for i, word := range words {
		if i == 0 {
			tmp = word
		} else if len(tmp)+len(word) < maxLength {
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
