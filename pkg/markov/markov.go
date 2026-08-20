// Package markovpkg builds order-N markov chains out of raw text, local files
// or web pages, and generates new sentences from them.
package markovpkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/PuerkitoBio/goquery"
)

// ErrInsufficientData is returned when the parsed corpus is too small to
// generate anything from.
var ErrInsufficientData = errors.New("not enough text to build a markov chain")

const (
	defaultOrder    = 2
	defaultMinWords = 8
	defaultMaxWords = 150
	// Discord rejects messages over 2000 characters, so stay well under it.
	defaultMaxChars = 1900

	// generateAttempts is how often Generate re-walks the chain looking for a
	// result that reaches minWords before settling for the longest walk it saw.
	generateAttempts = 12

	// prefixSeparator joins the words of a prefix into a map key. A NUL byte
	// never survives whitespace tokenization, so keys stay unambiguous.
	prefixSeparator = "\x00"

	// endMarker is recorded as the suffix of a segment's final prefix. Drawing
	// it ends the walk, which lets a sentence stop where a real message stopped
	// even when that message had no closing punctuation.
	endMarker = "\x00end"
)

// Chain is an order-N markov chain. The zero value is not usable; call New.
type Chain struct {
	order    int
	minWords int
	maxWords int
	maxChars int

	// states maps a prefix of `order` words to every word ever observed after
	// it. Repeats are kept, so picking uniformly reproduces source frequencies.
	states map[string][]string
	// starts holds the prefixes that opened a sentence. Walks begin at one of
	// these so output does not start mid-clause.
	starts []string
}

// Option customises a Chain built by New.
type Option func(*Chain)

// WithOrder sets how many words of context a prefix carries. Higher orders
// stay closer to the source text; lower orders wander more.
func WithOrder(order int) Option {
	return func(c *Chain) { c.order = order }
}

// WithMinWords sets the length a walk must reach before a terminal word is
// allowed to end it.
func WithMinWords(words int) Option {
	return func(c *Chain) { c.minWords = words }
}

// WithMaxWords caps how long a generated sentence may get.
func WithMaxWords(words int) Option {
	return func(c *Chain) { c.maxWords = words }
}

// WithMaxChars caps the generated sentence in characters.
func WithMaxChars(chars int) Option {
	return func(c *Chain) { c.maxChars = chars }
}

// New returns an empty chain ready to be fed text.
func New(opts ...Option) *Chain {
	c := &Chain{
		order:    defaultOrder,
		minWords: defaultMinWords,
		maxWords: defaultMaxWords,
		maxChars: defaultMaxChars,
		states:   make(map[string][]string),
	}
	for _, opt := range opts {
		opt(c)
	}

	// Clamp nonsensical options rather than panicking deep inside a walk.
	if c.order < 1 {
		c.order = defaultOrder
	}
	if c.minWords < c.order {
		c.minWords = c.order
	}
	if c.maxWords < c.minWords {
		c.maxWords = c.minWords
	}
	if c.maxChars < 1 {
		c.maxChars = defaultMaxChars
	}
	return c
}

// Parse folds one segment of text into the chain. A segment is a self-contained
// unit -- a single chat message, a paragraph, a document -- and the chain never
// walks from the tail of one segment into the head of another. Feeding several
// messages as one segment is what makes output read like two unrelated
// sentences welded together.
//
// Parse is additive: calling it repeatedly builds a single chain spanning the
// whole corpus. Recombination still happens across segments, but only where
// they genuinely share a prefix, which is the recombination worth having.
func (c *Chain) Parse(text string) {
	words := strings.Fields(text)
	if len(words) <= c.order {
		return
	}

	for i := 0; i+c.order < len(words); i++ {
		key := strings.Join(words[i:i+c.order], prefixSeparator)

		// The first prefix, and any prefix following a terminal word, is a
		// place a sentence can legitimately start.
		if i == 0 || util.IsTerminalWord(words[i-1]) {
			c.starts = append(c.starts, key)
		}
		c.states[key] = append(c.states[key], words[i+c.order])
	}

	// Record where the segment ran out, so a walk can stop here even if the
	// last word carries no punctuation.
	last := strings.Join(words[len(words)-c.order:], prefixSeparator)
	c.states[last] = append(c.states[last], endMarker)
}

// ParseSegments folds each segment into the chain independently. Empty and
// too-short segments are skipped.
func (c *Chain) ParseSegments(segments ...string) {
	for _, segment := range segments {
		c.Parse(segment)
	}
}

// Reset drops everything parsed so far.
func (c *Chain) Reset() {
	c.states = make(map[string][]string)
	c.starts = nil
}

// StateDictionary exposes the prefix -> suffixes map. Keys are the prefix words
// joined by prefixSeparator, and a suffix may be endMarker.
func (c *Chain) StateDictionary() map[string][]string {
	return c.states
}

// Generate walks the chain and returns a new sentence.
func (c *Chain) Generate() (string, error) {
	if len(c.starts) == 0 {
		return "", ErrInsufficientData
	}

	// Short or dead-end walks are common on small corpora, so retry a handful
	// of times before settling for the best one seen.
	var longest []string
	for range generateAttempts {
		words := c.walk()
		if len(words) >= c.minWords {
			return strings.Join(words, " "), nil
		}
		if len(words) > len(longest) {
			longest = words
		}
	}

	if len(longest) == 0 {
		return "", ErrInsufficientData
	}
	return strings.Join(longest, " "), nil
}

// walk produces one sentence, starting from a random sentence opening and
// stopping at a terminal word, a dead end, or one of the configured caps.
func (c *Chain) walk() []string {
	prefix := strings.Split(randomString(c.starts), prefixSeparator)

	words := make([]string, 0, c.maxWords)
	words = append(words, prefix...)
	chars := len(strings.Join(words, " "))

	for len(words) < c.maxWords {
		next := randomString(c.states[strings.Join(prefix, prefixSeparator)])
		if next == "" || next == endMarker {
			// Either a dead end, or a prefix that ended one of the segments.
			break
		}

		words = append(words, next)
		chars += len(next) + 1
		if chars >= c.maxChars {
			break
		}
		if len(words) >= c.minWords && util.IsTerminalWord(next) {
			break
		}

		// Slide the window: drop the oldest word, append the new one. prefix
		// has its own backing array, so this cannot corrupt words.
		copy(prefix, prefix[1:])
		prefix[len(prefix)-1] = next
	}

	return words
}

// ReadText parses text and immediately generates from it. Each line is treated
// as its own segment, so a newline-delimited corpus (a user's chat history, for
// instance) does not chain across message boundaries. Text without newlines is
// a single segment.
func (c *Chain) ReadText(text string) (string, error) {
	c.ParseSegments(strings.Split(text, "\n")...)
	return c.Generate()
}

// ReadFile parses the contents of filePath as a single segment and generates
// from it.
func (c *Chain) ReadFile(filePath string) (string, error) {
	text, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", filePath, err)
	}

	c.Parse(string(text))
	return c.Generate()
}

// ReadURL fetches a web page, parses its article text and generates from it.
func (c *Chain) ReadURL(rawURL string) (string, error) {
	if err := c.ParseURL(rawURL); err != nil {
		return "", err
	}
	return c.Generate()
}

const (
	fetchTimeout = 15 * time.Second
	// maxBodyBytes bounds how much of a remote page is read into memory.
	maxBodyBytes = 4 << 20
	// minStates is the point at which a selector is considered to have found
	// real body copy rather than a stray caption or cookie banner.
	minStates = 10
	userAgent = "copypastabot/2.0 (+https://github.com/stollenaar/copypastabotv2)"
)

var httpClient = &http.Client{Timeout: fetchTimeout}

// ParseURL fetches rawURL and folds its article text into the chain.
func (c *Chain) ParseURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	// The URL comes straight from a slash command, so refuse anything the
	// http client would treat as an exotic transport.
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", u.Host, err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("fetching %s: unexpected status %s", u.Host, res.Status)
	}
	if ct := res.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "html") {
		return fmt.Errorf("%s served %q, expected HTML", u.Host, ct)
	}

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(res.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", u.Host, err)
	}

	// Prefer real article copy, fall back to the main content area, then to any
	// paragraph on the page. Each <p> is parsed on its own so chains never run
	// across a paragraph boundary.
	selectors := []string{"article p", "main p", "p"}
	for i, selector := range selectors {
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			c.Parse(s.Text())
		})
		if len(c.states) >= minStates || i == len(selectors)-1 {
			break
		}
		c.Reset()
	}

	if len(c.starts) == 0 {
		return fmt.Errorf("%w: no readable article text at %s", ErrInsufficientData, u.Host)
	}
	return nil
}
