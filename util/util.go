package util

import (
	"os"
	"regexp"
)

var termRegex string

func IsTerminalWord(word string) bool {
	if termRegex == "" {
		termRegex = os.Getenv("TERMINAL_REGEX")

		// Setting the default regex
		if termRegex == "" {
			termRegex = `(\.|,|:|;|\?|!)$`
		}
	}
	compiled, _ := regexp.MatchString(termRegex, word)
	return compiled
}
