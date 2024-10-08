package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	sqsObject util.Object
)

func main() {
	lambda.Start(handler)
}

func handler(snsEvent events.SNSEvent) error {
	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &sqsObject)
	if err != nil {
		fmt.Println(err)
		return err
	}

	response := discordgo.WebhookEdit{
		Embeds: getCommandInformation(sqsObject.Command),
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Users: []string{},
			Roles: []string{},
		},
		Components: getSelectMenu(sqsObject.Command),
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return err
	}
	resp, err := util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, data)

	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString := string(bodyData)
		fmt.Println(resp, bodyString)
	}
	if err != nil {
		util.SendError(sqsObject)
		return err
	}
	return err
}

func getSelectMenu(command string) *[]discordgo.MessageComponent {
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select Command To Inspect",
					Options:     getSelectMenuOptions(command),
					CustomID:    "command_select",
				},
			},
		},
	}
	return &components
}

func getSelectMenuOptions(command string) (options []discordgo.SelectMenuOption) {
	for _, co := range util.Commands {
		if co.Description == "" {
			continue
		}
		option := discordgo.SelectMenuOption{
			Label: co.Name,
			Value: co.Name,
		}
		if command == co.Name {
			option.Default = true
		}
		options = append(options, option)
	}
	return
}

func getCommandInformation(command string) *[]*discordgo.MessageEmbed {
	embeds := []*discordgo.MessageEmbed{}

	if command == "" {
		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:       "Command Information",
			Description: "Select a command from the dropdown menu to learn more about",
		})
	} else {
		co := getCommand(command)
		if co == nil {
			embeds = append(embeds, &discordgo.MessageEmbed{
				Title:       "OOPSIE WHOOPSIE",
				Description: "An OOPSIE WHOOPSIE happened",
			})
			return &embeds
		}
		embed := discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s Command Information", toTitle(command)),
			Description: co.Description,
		}
		if len(co.Options) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s has the following options:", toTitle(command)),
			})
		}
		for _, option := range co.Options {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  toTitle(option.Name),
				Value: option.Description,
			})
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Value: fmt.Sprintf("Required: %t", option.Required),
			})
		}
		embeds = append(embeds, &embed)
	}
	return &embeds
}

func getCommand(command string) *discordgo.ApplicationCommand {
	for _, c := range util.Commands {
		if c.Name == command {
			return c
		}
	}
	return nil
}

func toTitle(in string) string {
	return cases.Title(language.AmericanEnglish, cases.Compact).String(in)
}
