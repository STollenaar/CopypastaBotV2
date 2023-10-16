package util

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	statsUtil "github.com/stollenaar/statisticsbot/util"
)

type KIND string

const (
	API_URL          = "https://discord.com/api/v10/%s/%s/%s"
	INTERACTION KIND = "interactions"
	WEBHOOK     KIND = "webhooks"

	POST_URL         = "https://discord.com/api/v10/interactions/%s/%s/callback"
	WEBHOOK_POST_URL = "https://discord.com/api/v10/webhooks/%s/%s"
	PATCH_URL        = "https://discord.com/api/v10/webhooks/%s/%s/messages/%s"
)

func SendError(sqsObject statsUtil.SQSObject) {
	e := "If you see this, and error likely happened. Whoops"
	response := discordgo.WebhookEdit{
		Content: &e,
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}
	SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, WEBHOOK, data)
}

func SendRequest(method, interactionID, interactionToken string, kind KIND, data []byte, messageID ...string) (*http.Response, error) {
	url := API_URL
	if kind == INTERACTION {
		url += "callback"
	} else {
		url += "/messages/"
		if len(messageID) == 0 {
			url += "@original"
		} else {
			url += messageID[0]
		}
	}
	req, err := http.NewRequest(method, fmt.Sprintf(url, kind, interactionID, interactionToken), bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	fmt.Println(*req)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return resp, nil
}

// Verifying the signature
func IsVerified(body, signature, timestamp string) bool {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		fmt.Println(fmt.Errorf("error decoding signature %w", err))
		return false
	}
	publicToken, err := ConfigFile.GetPublicDiscordToken()
	if err != nil {
		fmt.Println(fmt.Errorf("error fetching public key %w", err))
		return false
	}
	decodedKey, err := hex.DecodeString(publicToken)
	if err != nil {
		fmt.Println(fmt.Errorf("error decoding public key %w", err))
		return false
	}
	return ed25519.Verify(decodedKey, []byte(timestamp+body), decodedSig)
}
