package markovpkg

// GetMarkovURL creates a markov chain from the article text of a provided url.
func GetMarkovURL(url string) (string, error) {
	return New().ReadURL(url)
}
