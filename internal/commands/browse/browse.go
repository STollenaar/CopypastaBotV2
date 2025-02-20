package browse

import (
	"errors"
	"fmt"
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
			SubReddit: parsedArguments["SubReddit"],
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
		err := browser.Unmarshal([]byte(interaction.MessageComponentData().CustomID))
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
			Components: getActionRow(browser.Page, browser.SubReddit),
		}

		bot.InteractionResponseEdit(interaction.Interaction, &response)
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

func getActionRow(page int, subreddit string) *[]discordgo.MessageComponent {
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
