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
	response := util.ResponseObject{
		Data: discordgo.InteractionResponseData{
			Content: sqsObject.Data,
		},
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if sqsObject.Command == "markov" {
		return util.SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, data)
	} else {
		return errors.New("unimplemented route")
	}
}
