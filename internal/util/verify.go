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
	return ed25519.Verify([]byte(timestamp+body), decodedSig, []byte(ConfigFile.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN))
}
