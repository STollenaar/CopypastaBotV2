package main

import (
	"context"
	"encoding/json"
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

func (b *browserTracker) Unmarshal(data []byte) browserTracker {
	d := strings.Split(string(data), "|")
	page, _ := strconv.Atoi(d[1])
	return browserTracker{
		SubReddit: d[0],
		Page:      page,
		Action:    d[2],
	}
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

	// userID := interaction.Member.User.ID
	posts := util.DisplayRedditSubreddit(sqsObject.Data)
	embed := util.DisplayRedditPost(posts[0].ID, true)[0]

	response := discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: getActionRow(0, sqsObject.Data),
	}
	data, _ := json.Marshal(response)

	return util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, data)
}

func getActionRow(page int, subreddit string) []discordgo.MessageComponent {
	browser := browserTracker{
		SubReddit: subreddit,
		Page:      page,
		Action:    "previous",
	}
	bdataPrev := browser.Marshal()
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
