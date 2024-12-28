package markovpkg

// GetUserMarkov create a markov chain from an user
func GetUserMarkov(data string) (string, error) {
	markov := New()

	generated := markov.ReadText(data)
	return generated, nil
}
