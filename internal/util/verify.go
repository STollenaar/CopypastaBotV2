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
	decodedKey, err := hex.DecodeString(ConfigFile.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN)
	if err != nil {
		fmt.Println(fmt.Errorf("error decoding public key %w", err))
		return false
	}
	return ed25519.Verify([]byte(timestamp+body), decodedSig, decodedKey)
}
