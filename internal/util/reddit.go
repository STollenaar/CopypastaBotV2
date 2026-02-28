package util

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var (
	images       []string
	videos       []string
	redditClient *reddit.Client

	re *regexp.Regexp
)

func init() {
	images = []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".gif"}
	videos = []string{"youtube", "gfycat", "youtu"}

	r, err := regexp.Compile(ConfigFile.TERMINAL_REGEX)
	if err != nil {
		slog.Error("Error compiling terminal regex", slog.Any("err", err))
	}
	re = r
}

func createRedditClient() {
	clientId, _ := ConfigFile.GetRedditClientID()
	secret, _ := ConfigFile.GetRedditClientSecret()
	username, _ := ConfigFile.GetRedditUsername()
	password, _ := ConfigFile.GetRedditPassword()

	r, err := reddit.NewClient(reddit.Credentials{
		ID:       clientId,
		Secret:   secret,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln(fmt.Errorf("failure initializing reddit client %w", err))
	}
	redditClient = r
}

func IsTerminalWord(word string) bool {
	return re.MatchString(word)
}

func DisplayRedditSubreddit(subreddit string) []*reddit.Post {
	if redditClient == nil {
		createRedditClient()
	}
	posts, _, err := redditClient.Subreddit.HotPosts(context.TODO(), subreddit, &reddit.ListOptions{})
	if err != nil {
		slog.Error("Error fetching hot posts for subreddit", slog.String("subreddit", subreddit), slog.Any("err", err))
	}

	return posts
}

func GetRedditPost(redditPostID string) *reddit.PostAndComments {
	if redditClient == nil {
		createRedditClient()
	}
	postCommnents, _, err := redditClient.Post.Get(context.TODO(), redditPostID)
	if err != nil {
		slog.Error("Error fetching reddit post", slog.String("postID", redditPostID), slog.Any("err", err))
	}

	return postCommnents
}

func DisplayRedditPost(redditPostID string, singleEmbed bool) (embeds []discord.Embed) {
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
		embed := discord.Embed{}

		// Making sure if it's not a text content that the embed is set correctly
		if uri, err := url.ParseRequestURI(content); err == nil {
			if arrayContainsSub(images, uri.RequestURI()) {
				embed.Image = &discord.EmbedResource{
					URL: content,
				}
			} else if arrayContainsSub(videos, uri.RequestURI()) {
				embed.Video = &discord.EmbedResource{
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
			if len(embed.Title) >= 256 {
				embed.Title = embed.Title[:252] + "..."
			}
			embed.Author = &discord.EmbedAuthor{
				Name: postCommnents.Post.Author,
				URL:  ("https://www.reddit.com/u/" + postCommnents.Post.Author),
			}
		}

		// Only adding the footer if this is the last entry
		if i == len(contents)-1 {
			embed.Fields = []discord.EmbedField{
				{
					Name:   "Reddit Thread:",
					Value:  ("https://www.reddit.com" + postCommnents.Post.Permalink),
					Inline: Pointer(true),
				},
			}
			embed.Footer = &discord.EmbedFooter{
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
