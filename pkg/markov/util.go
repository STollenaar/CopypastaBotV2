package markovpkg

import (
	"math/rand/v2"
)

// randomString returns a uniformly random element of slice, or "" when the
// slice is empty. Callers rely on that empty string to spot a dead end, which
// is unambiguous because whitespace tokenization never yields an empty word.
func randomString(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[rand.IntN(len(slice))]
}
