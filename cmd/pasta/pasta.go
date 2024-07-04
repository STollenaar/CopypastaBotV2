package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/bwmarrin/discordgo"
)

func main() {
	lambda.Start(handler)
}

func handler(snsEvent events.SNSEvent) error {
	var req events.APIGatewayProxyRequest
	var interaction discordgo.Interaction
	var response discordgo.WebhookEdit

	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &req)
	if err !=nil {
		return err
	}
	err = json.Unmarshal([]byte(req.Body), &interaction)
	if err !=nil {
		return err
	}
	
	fmt.Println(interaction.ApplicationCommandData().Options)
	parsedArguments := util.ParseArguments([]string{"url", "postid"}, interaction.ApplicationCommandData().Options)
	fmt.Println(parsedArguments)


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
		response.Embeds = &embeds
		response.Content = aws.String("")
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = util.SendRequest("PATCH", interaction.AppID, interaction.Token, util.WEBHOOK, data)

	return err
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
