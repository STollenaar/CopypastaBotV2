package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

// MessageObject general messageobject for functions
type MessageObject struct {
	GuildID   string               `bson:"GuildID" json:"GuildID"`
	ChannelID string               `bson:"ChannelID" json:"ChannelID"`
	MessageID string               `bson:"_id" json:"MessageID"`
	Author    string               `bson:"Author" json:"Author"`
	Content   []string             `bson:"Content" json:"Content"`
	Date      discordgo.TimeStamps `bson:"Date" json:"Date"`
}

func GetMessageObject(object statsUtil.SQSObject) (discordgo.Message, error) {
	if _, err := strconv.Atoi(object.Token); err == nil {
		return discordgo.Message{}, errors.New("object token is a snowflake")
	}
	resp, err := SendRequest("GET", object.ApplicationID, object.Token, WEBHOOK, []byte{})
	var bodyString string
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString = string(bodyData)
		fmt.Println(resp, bodyString)
	}
	if err != nil {
		return discordgo.Message{}, fmt.Errorf("error sending request %v", err)
	}
	var message discordgo.Message
	err = json.Unmarshal([]byte(bodyString), &message)
	if err != nil {
		return discordgo.Message{}, fmt.Errorf("error parsing into interaction with data: %s, and error: %v", bodyString, err)
	}
	return message, nil
}
