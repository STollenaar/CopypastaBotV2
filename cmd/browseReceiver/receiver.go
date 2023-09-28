package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

type browserTracker struct {
	SubReddit string `json:"subReddit"`
	Page      int    `json:"page"`
	Action    string `json:"action"`
}

func (b *browserTracker) Marshal() string {
	return fmt.Sprintf("%s|%d|%s", b.SubReddit, b.Page, b.Action)
}

func (b *browserTracker) Unmarshal(data []byte) error {
	d := strings.Split(string(data), "|")
	if len(d) != 3 {
		return errors.New("unknown data format")
	}
	page, _ := strconv.Atoi(d[1])
	b.SubReddit = d[0]
	b.Page = page
	b.Action = d[2]
	return nil
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	var sqsObject statsUtil.SQSObject
	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), &sqsObject)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var browser browserTracker
	err = browser.Unmarshal([]byte(sqsObject.Data))
	if err != nil {
		browser = browserTracker{
			SubReddit: sqsObject.Data,
			Page:      0,
			Action:    "",
		}
	}

	// userID := interaction.Member.User.ID
	posts := util.DisplayRedditSubreddit(browser.SubReddit)
	embed := util.DisplayRedditPost(posts[browser.Page].ID, true)[0]

	response := discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: getActionRow(browser.Page, browser.SubReddit),
	}
	data, _ := json.Marshal(response)

	return util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, data)
}

func getActionRow(page int, subreddit string) []discordgo.MessageComponent {
	browser := browserTracker{
		SubReddit: subreddit,
		Page:      page - 1,
		Action:    "previous",
	}
	if browser.Page < 0 {
		browser.Page = 0
	}

	bdataPrev := browser.Marshal()
	browser.Page = page + 1
	browser.Action = "next"
	bdataNext := browser.Marshal()

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: bdataPrev,
					Label:    "Previous",
				},
				discordgo.Button{
					CustomID: bdataNext,
					Label:    "Next",
				},
			},
		},
	}
}
