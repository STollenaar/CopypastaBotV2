package util

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"
)

const (
	URL = "https://discord.com/api/v10/interactions/%s/%s/callback"
)

func SendRequest(interactionID, interactionToken string, data []byte) {
	// Create a HTTP post request
	req, err := http.NewRequest("POST", fmt.Sprintf(URL, interactionID, interactionToken), bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
}

// Verifying the signature
func IsVerified(body, signature, timestamp string) bool {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		fmt.Println(fmt.Errorf("error decoding signature %w", err))
		return false
	}
	decodedKey, err := hex.DecodeString(ConfigFile.GetPublicDiscordToken())
	if err != nil {
		fmt.Println(fmt.Errorf("error decoding public key %w", err))
		return false
	}
	return ed25519.Verify(decodedKey, []byte(timestamp+body), decodedSig)
}
