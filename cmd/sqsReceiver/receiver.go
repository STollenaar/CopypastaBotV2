package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	"github.com/stollenaar/copypastabotv2/pkg/markov"
)

var (
	sendTimeout = true
	sqsObject   util.Object
)

func main() {
	lambda.StartWithOptions(handler, lambda.WithEnableSIGTERM(timeoutHandler))
}

func timeoutHandler() {
	if sendTimeout {
		lsm := os.Getenv("LogStreamName")
		d := fmt.Sprintf("If you see this, and error likely happened. Give this to the distinguished Copypastabot Engineer: %s", lsm)
		response := discordgo.WebhookEdit{
			Content: &d,
		}
		data, _ := json.Marshal(response)
		util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, data)
	}
}

func handler(snsEvent events.SNSEvent) error {

	err := json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &sqsObject)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if _, snerr := strconv.Atoi(sqsObject.Token); snerr != nil && sqsObject.Command == "speak" {
		d := "Data received, building Markov"
		response := discordgo.WebhookEdit{
			Content: &d,
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
			fmt.Println(err)
		}
	}

	var markovData string

	switch sqsObject.Type {
	case "url":
		markovData = handleURL(sqsObject.Data)
	case "user":
		markovData = handleUser(sqsObject.Data)
	default:
		return fmt.Errorf("unimplemented type: %s", sqsObject.Type)
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
		resp, err := util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, data)
		if resp != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			bodyData := buf.String()

			bodyString := string(bodyData)
			fmt.Println(resp, bodyString)
		}
		return err
	case "speak":
		sqsMessage := util.Object{
			Token:         sqsObject.Token,
			Type:          sqsObject.Type,
			Command:       "speak",
			GuildID:       sqsObject.GuildID,
			ApplicationID: sqsObject.ApplicationID,
			Data:          markovData,
		}

		sqsMessageData, _ := json.Marshal(sqsMessage)
		err := util.PublishObject("speakReceiver", string(sqsMessageData))
		if _, snerr := strconv.Atoi(sqsObject.Token); snerr != nil && err != nil {
			sendTimeout = false
			fmt.Println(err)
			resp, err := util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, []byte(err.Error()))
			if resp != nil {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				bodyData := buf.String()

				bodyString := string(bodyData)
				fmt.Println(resp, bodyString)
			}
			return err
		}
		return err
	default:
		return fmt.Errorf("unimplemented command: %s", sqsObject.Command)
	}
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
