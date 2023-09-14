package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

var (
	Bot         *discordgo.Session
	statsbotUrl string
	reTarget    *regexp.Regexp
)

func init() {
	reTarget = regexp.MustCompile(`[\<>@#&!]`)
}

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := discordgo.InteractionResponseData{
		Content: "Pong",
	}
	data, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}

// Command create a markov chain from an URL
func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if Bot == nil {
		Bot = bot
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading Markov...",
		},
	})

	parsedArguments := util.ParseArguments([]string{"url", "user"}, interaction)

	if url, ok := parsedArguments["Url"]; ok {
		MarkovURLCommand(interaction, url)
	} else if user, ok := parsedArguments["User"]; ok {
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")
		MarkovUserCommand(interaction, user)
	} else {
		unknownState := "Unknown state entered"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &unknownState,
		})
	}
}
