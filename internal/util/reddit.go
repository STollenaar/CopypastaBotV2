package util

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	images       []string
	videos       []string
	redditClient *reddit.Client
)

func init() {
	images = []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".gif"}
	videos = []string{"youtube", "gfycat", "youtu"}
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
	compiled, err := regexp.MatchString(ConfigFile.TERMINAL_REGEX, word)
	if err != nil {
		fmt.Println(fmt.Errorf("error matching string with regex %s, on word %s. %w", ConfigFile.TERMINAL_REGEX, word, err))
	}
	return compiled
}

func DisplayRedditSubreddit(subreddit string) []*reddit.Post {
	if redditClient == nil {
		createRedditClient()
	}
	posts, _, err := redditClient.Subreddit.HotPosts(context.TODO(), subreddit, &reddit.ListOptions{})
	if err != nil {
		fmt.Println(fmt.Errorf("error fetching hot posts for subreddit %s. With error: %w", subreddit, err))
	}

	return posts
}

func GetRedditPost(redditPostID string) *reddit.PostAndComments {
	if redditClient == nil {
		createRedditClient()
	}
	postCommnents, _, err := redditClient.Post.Get(context.TODO(), redditPostID)
	if err != nil {
		fmt.Println(fmt.Errorf("error fetching from reddit with error: %w", err))
	}

	return postCommnents
}

func DisplayRedditPost(redditPostID string, singleEmbed bool) (embeds []*discordgo.MessageEmbed) {
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
			if len(embed.Title) >= 256 {
				embed.Title = embed.Title[:252] + "..."
			}
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
		embeds = append(embeds, &embed)
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

// CommandParsed parsed struct for count command
type CommandParsed map[string]string

func ParseArguments(arguments []string, options []*discordgo.ApplicationCommandInteractionDataOption) (parsedArguments CommandParsed) {
	parsedArguments = make(map[string]string)
	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	for _, arg := range arguments {
		if option, ok := optionMap[arg]; ok {
			parsedArguments[cases.Title(language.English).String(arg)] = option.Value.(string)
		}
	}

	return parsedArguments
}
