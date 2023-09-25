package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	"github.com/stollenaar/copypastabotv2/pkg/markov"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

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

	var markovData string

	switch sqsObject.Type {
	case "url":
		markovData = handleURL(sqsObject.Data)
	case "user":
		markovData = handleUser(sqsObject.Data)
	default:
		return errors.New("unimplemented type")
	}

	switch sqsObject.Command {
	case "markov":
		response := discordgo.WebhookEdit{
			Content: &markovData,
		}

		data, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, data)
	case "speak":
	default:
		return errors.New("unimplemented command")
	}
	return errors.New("how did we get here?")
}

func handleURL(input string) string {
	data, err := markov.GetMarkovURL(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}

func handleUser(input string) string {
	data, err := markov.GetUserMarkov(input)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return data
}
