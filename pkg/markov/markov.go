package markovpkg

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/PuerkitoBio/goquery"
)

type markov struct {
	states map[[2]string][]string
}

func New() *markov {
	return &markov{}
}

func (m *markov) ReadText(text string) string {
	m.parse(text)

	return m.generate()
}

func (m *markov) ReadFile(filePath string) string {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read data from the file
	text, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	m.parse(string(text))

	return m.generate()
}

func (m *markov) ReadURL(URL string) (string, error) {
	// Open web page
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		log.Fatal(err)
	}

	// Search for <p></p> under <article></article> tags
	doc.Find("article").Each(func(i int, s *goquery.Selection) {
		text := s.Find("p").Text()
		m.parse(text)
	})
	if len(m.states) < 10 {
		return "", errors.New("article from URL couldn't be parsed")
	}

	return m.generate(), nil
}

func (m *markov) StateDictionary() map[[2]string][]string {
	return m.states
}

// Parse input text into states map
func (m *markov) parse(text string) {
	// Initialize markov.states map
	m.states = make(map[[2]string][]string)

	words := strings.Split(text, " ")

	for i := 0; i < len(words)-2; i++ {
		// Initialize prefix with two consecutive words as the key
		prefix := [2]string{words[i], words[i+1]}

		// Assign the third word as value to the prefix
		m.states[prefix] = append(m.states[prefix], words[i+2])
	}
}

// Generate markov senetence based on a given length
func (m *markov) generate() string {
	var sentence bytes.Buffer

	// Initialize prefix with a random key
	prefix := m.getRandomPrefix([2]string{"", ""})
	sentence.WriteString(strings.Join(prefix[:], " ") + " ")

	for {
		suffix := getRandomWord(m.states[prefix])
		sentence.WriteString(suffix + " ")

		// Break the loop if suffix ends in "." and sentenceLength is enough
		if util.IsTerminalWord(suffix) {
			break
		}

		prefix = [2]string{prefix[1], suffix}
	}

	return sentence.String()
}

// Return a random prefix other than the one in the arguments
func (m *markov) getRandomPrefix(prefix [2]string) [2]string {
	// By default, Go orders keys randomly for maps
	for key := range m.states {
		if key != prefix {
			prefix = key
			break
		}
	}

	return prefix
}
