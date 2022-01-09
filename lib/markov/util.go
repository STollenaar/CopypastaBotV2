package markov

import (
	"math/rand"
)

// Return a random element from a given string slice
func getRandomWord(slice []string) string {
	if cap(slice) != 0 {
		return slice[rand.Intn(len(slice))]
	} else {
		return ""
	}
}
