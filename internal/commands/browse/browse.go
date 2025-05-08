package browse

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

type browserTracker struct {
	UserId    string `json:"userId"`
	SubReddit string `json:"subReddit"`
	Page      int    `json:"page"`
	Action    string `json:"action"`
}

func (b *browserTracker) Marshal() string {
	return fmt.Sprintf("%s|%d|%s|%s", b.SubReddit, b.Page, b.Action, b.UserId)
}

func (b *browserTracker) Unmarshal(data []byte) error {
	d := strings.Split(string(data), "|")
	if len(d) != 4 {
		return errors.New("unknown data format")
	}
	page, _ := strconv.Atoi(d[1])
	b.SubReddit = d[0]
	b.Page = page
	b.Action = d[2]
	b.UserId = d[3]
	return nil
}

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var parsedArguments util.CommandParsed
	var browser browserTracker
	if interaction.Type == discordgo.InteractionType(discordgo.InteractionApplicationCommand) {
		bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Loading...",
			},
		})
		parsedArguments = util.ParseArguments([]string{"subreddit"}, interaction.ApplicationCommandData().Options)
		browser = browserTracker{
			SubReddit: parsedArguments["Subreddit"],
			Page:      0,
			UserId:    interaction.Member.User.ID,
		}
	} else {
		bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
			Data: &discordgo.InteractionResponseData{
				Content: "Loading...",
			},
		})
		err := browser.Unmarshal([]byte(interaction.Interaction.MessageComponentData().CustomID))
		if err != nil {
			fmt.Printf("Error unmarshalling browser data: %v\n", err)
			return
		}
	}

	userID := interaction.Member.User.ID
	if userID == browser.UserId {
		posts := util.DisplayRedditSubreddit(browser.SubReddit)
		embed := util.DisplayRedditPost(posts[browser.Page].ID, true)[0]

		response := discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: getActionRow(browser),
		}

		_, err := bot.InteractionResponseEdit(interaction.Interaction, &response)
		if err != nil {
			log.Println(err)
		}
	}
}

// func lambdaHandler(snsEvent events.SNSEvent) error {
// 	var interaction discordgo.Interaction
// 	var response discordgo.WebhookEdit

// 	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &req)
// 	if err != nil {
// 		return err
// 	}
// 	err = json.Unmarshal([]byte(req.Body), &interaction)
// 	if err != nil {
// 		return err
// 	}

// 	parsedArguments := util.ParseArguments([]string{"name"}, interaction.ApplicationCommandData().Options)
// 	if parsedArguments["Name"] == "" {
// 		parsedArguments["Name"] = "all"
// 	}

// 	sqsMessage := util.Object{
// 		Token:         interaction.Token,
// 		Command:       "browse",
// 		GuildID:       interaction.GuildID,
// 		ApplicationID: interaction.AppID,
// 		Data:          parsedArguments["Name"],
// 	}
// 	sqsMessageData, _ := json.Marshal(sqsMessage)
// 	err = util.PublishObject("browseReceiver", string(sqsMessageData))
// 	if err != nil {
// 		fmt.Printf("Encountered an error while processing the browse command: %v\n", err)
// 		return err
// 	}

// 	data, _ := json.Marshal(response)
// 	fmt.Printf("Responding with %s\n", string(data))
// 	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)
// 	return err
// }

func getActionRow(browser browserTracker) *[]discordgo.MessageComponent {
	prevBrowser := browserTracker{
		UserId:    browser.UserId,
		SubReddit: browser.SubReddit,
		Page:      browser.Page - 1,
		Action:    "previous",
	}

	nextBrowser := browserTracker{
		UserId:    browser.UserId,
		SubReddit: browser.SubReddit,
		Page:      browser.Page + 1,
		Action:    "next",
	}

	if prevBrowser.Page < 0 {
		prevBrowser.Page = 0
	}

	bdataPrev := prevBrowser.Marshal()
	bdataNext := nextBrowser.Marshal()

	return &[]discordgo.MessageComponent{
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
