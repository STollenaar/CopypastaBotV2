package markovpkg

import "strings"

// GetUserMarkov creates a markov chain from a user's messages. Each message is
// its own segment, and embedded newlines split further, so the chain never runs
// from the end of one message into the start of the next. Passing a single
// pre-joined blob still works; it is simply treated as one segment.
func GetUserMarkov(messages ...string) (string, error) {
	c := New()
	for _, message := range messages {
		c.ParseSegments(strings.Split(message, "\n")...)
	}
	return c.Generate()
}
