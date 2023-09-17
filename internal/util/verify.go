package util

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

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
