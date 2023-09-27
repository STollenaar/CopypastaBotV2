package util

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"
)

const (
	POST_URL  = "https://discord.com/api/v10/interactions/%s/%s/callback"
	PATCH_URL = "https://discord.com/api/v10/webhooks/%s/%s/messages/%s"
)

func SendRequest(method, interactionID, interactionToken string, data []byte, messageID ...string) error {
	if len(messageID) == 0 {
		messageID = append(messageID, "@original")
	}
	// Create a HTTP post request
	var req *http.Request
	var err error
	switch method {
	case "POST":
		req, err = http.NewRequest("POST", fmt.Sprintf(POST_URL, interactionID, interactionToken), bytes.NewBuffer(data))
	case "PATCH":
		req, err = http.NewRequest("PATCH", fmt.Sprintf(PATCH_URL, interactionID, interactionToken, messageID[0]), bytes.NewBuffer(data))
	}
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		fmt.Println(err)
		return err
	}
	client := &http.Client{}
	fmt.Println(*req)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	bodyData := buf.String()

	bodyString := string(bodyData)
	fmt.Println(resp, bodyString)
	return nil
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
