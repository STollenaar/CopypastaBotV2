package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

var (
	//go:embed chatRole.txt
	systemPrompt string
	//go:embed chatRoleSpeak.txt
	systemPromptSpeak string
	//go:embed cavemanRole.txt
	systemCaveman string

	sqsClient   *sqs.Client
	sendTimeout = true
	sqsObject   statsUtil.SQSObject
)

func init() {
	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Error loading AWS config:", err)
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
	message, err := util.GetMessageObject(sqsObject)
	if err != nil {
		fmt.Println(err)
		util.SendError(sqsObject)
		return err
	}
	var prompt string
	switch sqsObject.Command {
	case "caveman":
		fallthrough
	case "caveman-vc":
		prompt = systemCaveman
	case "respond":
		fallthrough
	case "chat":
		prompt = systemPrompt
	case "respond-vc":
		fallthrough
	case "speak":
		prompt = systemPromptSpeak
	}

	chatRSP, err := util.GetChatGPTResponse(prompt, sqsObject.Data, message.Interaction.User.ID)
	if err != nil {
		fmt.Println(err)
		e := "If you see this, and error likely happened. Whoops"
		response := discordgo.WebhookEdit{
			Content: &e,
		}

		data, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
			return err
		}
		_, err = util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, data)
		return err
	}

	switch sqsObject.Command {
	case "respond":
		fallthrough
	case "caveman":
		fallthrough
	case "chat":
		// Getting around the 4096 word limit
		contents := util.BreakContent(chatRSP.Choices[0].Message.Content, 4096)
		var embeds []*discordgo.MessageEmbed
		for _, content := range contents {
			embed := discordgo.MessageEmbed{}
			embed.Description = content
			embeds = append(embeds, &embed)
		}
		response := discordgo.WebhookEdit{
			Embeds: &embeds,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Users: []string{},
				Roles: []string{},
			},
		}

		data, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
			return err
		}
		resp, err := util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, util.WEBHOOK, data)
		sendTimeout = false

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
	case "respond-vc":
		fallthrough
	case "caveman-vc":
		fallthrough
	case "speak":
		sqsObject.Data = chatRSP.Choices[0].Message.Content
		sqsMessageData, _ := json.Marshal(sqsObject)
		_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(sqsMessageData)),
			QueueUrl:    aws.String(util.ConfigFile.AWS_SQS_URL),
		})
		if err != nil {
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
