package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	d, _ := json.Marshal(req)
	fmt.Println(string(d))
	var interaction discordgo.Interaction

	json.Unmarshal([]byte(req.Body), &interaction)
	fmt.Println(interaction.ApplicationCommandData().Options)
	parsedArguments := util.ParseArguments([]string{"url", "postid"}, interaction.ApplicationCommandData().Options)
	fmt.Println(parsedArguments)
	response := util.ResponseObject{
		Data: discordgo.InteractionResponseData{
			Content: fmt.Sprintln("something went wrong"),
		},
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	}

	if parsedArguments["Url"] != "" {
		if uri, err := url.ParseRequestURI(parsedArguments["Url"]); err == nil {
			path := strings.Split(uri.Path, "/")
			postID := findIndex(path, "comments")
			if postID != -1 {
				parsedArguments["Postid"] = path[postID+1]
			}
			parsedArguments["Postid"] = path[postID+1]
		}
	}
	if parsedArguments["Postid"] != "" {
		embeds := util.DisplayRedditPost(parsedArguments["Postid"], false)
		response.Data.Embeds = embeds
		response.Data.Content = ""
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
