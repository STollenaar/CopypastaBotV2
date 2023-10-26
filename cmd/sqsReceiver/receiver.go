package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	"github.com/stollenaar/copypastabotv2/pkg/markov"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

var (
	sqsClient   *sqs.Client
	sendTimeout = true
	sqsObject   statsUtil.SQSObject
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	sqsClient = sqs.NewFromConfig(cfg)
}

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

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), &sqsObject)
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
		sqsMessage := statsUtil.SQSObject{
			Token:         sqsObject.Token,
			Type:          sqsObject.Type,
			Command:       "speak",
			GuildID:       sqsObject.GuildID,
			ApplicationID: sqsObject.ApplicationID,
			Data:          markovData,
		}

		sqsMessageData, _ := json.Marshal(sqsMessage)
		_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(sqsMessageData)),
			QueueUrl:    aws.String(util.ConfigFile.AWS_SQS_URL),
		})
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
