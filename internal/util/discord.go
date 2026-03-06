package util

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/disgoorg/disgo/discord"
)

type KIND string

const (
	API_URL          = "https://discord.com/api/v10/%s/%s/%s"
	INTERACTION KIND = "interactions"
	WEBHOOK     KIND = "webhooks"
	MESSAGE     KIND = "message"

	POST_URL         = "https://discord.com/api/v10/interactions/%s/%s/callback"
	WEBHOOK_POST_URL = "https://discord.com/api/v10/webhooks/%s/%s"
	PATCH_URL        = "https://discord.com/api/v10/webhooks/%s/%s/messages/%s"
	MESSAGE_URL      = "https://discord.com/api/v10/channels/%s/messages"
)

func SendError(sqsObject Object) {
	e := "If you see this, and error likely happened. Whoops"
	response := discord.MessageUpdate{
		Content: &e,
	}

	data, err := json.Marshal(response)
	if err != nil {
		slog.Error("Error marshalling error response", slog.Any("err", err))
	}
	_, err = SendRequest("PATCH", sqsObject.ApplicationID, sqsObject.Token, WEBHOOK, data)
	if err != nil {
		slog.Error("Error sending patch", slog.Any("err", err))
	}
}

func SendRequest(method, interactionID, interactionToken string, kind KIND, data []byte, messageID ...string) (*http.Response, error) {
	url := API_URL
	var err error
	var req *http.Request

	switch kind {
	case INTERACTION:
		url += "callback"
		url = fmt.Sprintf(url, kind, interactionID, interactionToken)
	case MESSAGE:
		url = MESSAGE_URL
		url = fmt.Sprintf(url, interactionID)
	default:
		url += "/messages/"
		if len(messageID) == 0 {
			url += "@original"
		} else {
			url += messageID[0]
		}
		url = fmt.Sprintf(url, kind, interactionID, interactionToken)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))

	if err != nil {
		slog.Error("Error creating HTTP request", slog.Any("err", err))
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	token := ConfigFile.GetDiscordToken()
	req.Header.Add("Authorization", "Bot "+token)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending HTTP request", slog.Any("err", err))
		return nil, err
	}
	return resp, nil
}

// Verifying the signature
func IsVerified(body, signature, timestamp string) bool {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		slog.Error("Error decoding signature", slog.Any("err", err))
		return false
	}
	publicToken, err := ConfigFile.GetPublicDiscordToken()
	if err != nil {
		slog.Error("Error fetching public key", slog.Any("err", err))
		return false
	}
	decodedKey, err := hex.DecodeString(publicToken)
	if err != nil {
		slog.Error("Error decoding public key", slog.Any("err", err))
		return false
	}
	return ed25519.Verify(decodedKey, []byte(timestamp+body), decodedSig)
}

func SendAsWebhook(data []byte) (*http.Response, error) {

	webhook, err := ConfigFile.GetDiscordWebhook()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(WEBHOOK_POST_URL, webhook.id, webhook.token), bytes.NewBuffer(data))

	if err != nil {
		slog.Error("Error creating webhook request", slog.Any("err", err))
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	slog.Debug("Sending webhook request", slog.Any("request", *req))
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending webhook request", slog.Any("err", err))
		return nil, err
	}
	return resp, nil
}
