package markovpkg

import (
	URL "net/url"
)

// GetMarkovURL creating a markov chain from a provided url
func GetMarkovURL(url string) (string, error) {
	u, err := URL.ParseRequestURI(url)

	if err != nil {
		return "", err
	}
	markov := New()

	return markov.ReadURL(u.String())
}
