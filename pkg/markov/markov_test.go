package markovpkg

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const corpus = "the quick brown fox jumps over the lazy dog. " +
	"the quick brown fox sleeps beside the lazy cat. " +
	"a lazy dog naps while the quick brown fox watches the door."

func TestParseIsAdditive(t *testing.T) {
	c := New()
	c.Parse("one two three four five six.")
	first := len(c.StateDictionary())

	c.Parse("seven eight nine ten eleven twelve.")
	if second := len(c.StateDictionary()); second <= first {
		t.Fatalf("second Parse did not add states: had %d, now %d", first, second)
	}
}

func TestGenerateProducesWordsFromCorpus(t *testing.T) {
	c := New()
	got, err := c.ReadText(corpus)
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}
	if got == "" {
		t.Fatal("ReadText returned an empty sentence")
	}

	known := make(map[string]bool)
	for _, w := range strings.Fields(corpus) {
		known[w] = true
	}
	for _, w := range strings.Fields(got) {
		if !known[w] {
			t.Errorf("generated word %q is not in the corpus", w)
		}
	}
}

// A corpus whose every prefix is unique gives the walk nowhere to go but the
// end of the text. The old implementation looped forever there.
func TestGenerateTerminatesOnDeadEnd(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		c := New()
		if _, err := c.ReadText("alpha beta gamma delta epsilon zeta eta theta"); err != nil {
			t.Errorf("ReadText: %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Generate did not terminate on a dead-end chain")
	}
}

func TestGenerateRespectsMaxWords(t *testing.T) {
	// "spam spam spam ..." makes one self-referential prefix with no exit.
	c := New(WithMaxWords(20), WithMinWords(5))
	got, err := c.ReadText(strings.Repeat("spam ", 200))
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}
	if n := len(strings.Fields(got)); n > 20 {
		t.Fatalf("generated %d words, want at most 20", n)
	}
}

func TestGenerateRespectsMaxChars(t *testing.T) {
	c := New(WithMaxChars(60), WithMinWords(2), WithMaxWords(500))
	got, err := c.ReadText(strings.Repeat("supercalifragilistic ", 200))
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}
	if len(got) > 90 { // maxChars plus at most one overshooting word
		t.Fatalf("generated %d chars, want roughly at most 60: %q", len(got), got)
	}
}

func TestGenerateEmptyCorpus(t *testing.T) {
	c := New()
	if _, err := c.ReadText("   "); !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("got %v, want ErrInsufficientData", err)
	}
}

func TestOrderOption(t *testing.T) {
	c := New(WithOrder(3))
	c.Parse("a b c d")
	for key := range c.StateDictionary() {
		if n := len(strings.Split(key, prefixSeparator)); n != 3 {
			t.Fatalf("prefix %q has %d words, want 3", key, n)
		}
	}
}

func TestWalkStartsAtSentenceBoundary(t *testing.T) {
	c := New()
	c.Parse("first second third fourth. alpha beta gamma delta.")

	starts := map[string]bool{"first\x00second": true, "alpha\x00beta": true}
	for range 50 {
		words := c.walk()
		if got := strings.Join(words[:2], prefixSeparator); !starts[got] {
			t.Fatalf("walk started at %q, want a sentence opening", got)
		}
	}
}

func TestParseURLPrefersArticleText(t *testing.T) {
	page := `<html><body>
		<nav><p>home about contact</p></nav>
		<article><p>` + corpus + `</p><p>` + corpus + `</p></article>
	</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	}))
	defer srv.Close()

	c := New()
	if err := c.ParseURL(srv.URL); err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if _, ok := c.StateDictionary()["home"+prefixSeparator+"about"]; ok {
		t.Error("nav text leaked into the chain despite an <article> being present")
	}
}

func TestParseURLErrors(t *testing.T) {
	statusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer statusSrv.Close()

	jsonSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"a":1}`))
	}))
	defer jsonSrv.Close()

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><div>no paragraphs here</div></body></html>`))
	}))
	defer emptySrv.Close()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"not a url", "definitely not a url", "invalid url"},
		{"bad scheme", "file:///etc/passwd", "unsupported url scheme"},
		{"non 2xx", statusSrv.URL, "unexpected status"},
		{"not html", jsonSrv.URL, "expected HTML"},
		{"no article text", emptySrv.URL, "no readable article text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New().ParseURL(tt.url)
			if err == nil {
				t.Fatalf("ParseURL(%q) succeeded, want error", tt.url)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestSegmentsDoNotChainIntoEachOther(t *testing.T) {
	c := New(WithMinWords(2), WithMaxWords(50))
	c.ParseSegments("alpha beta gamma delta", "one two three four")

	// The segments share no prefix, so no walk may cross between them.
	segmentOf := map[string]int{"alpha": 1, "beta": 1, "gamma": 1, "delta": 1}
	for _, w := range []string{"one", "two", "three", "four"} {
		segmentOf[w] = 2
	}

	for range 200 {
		words := c.walk()
		first := segmentOf[words[0]]
		for _, w := range words {
			if segmentOf[w] != first {
				t.Fatalf("walk crossed a segment boundary: %q", strings.Join(words, " "))
			}
		}
	}
}

func TestReadTextTreatsLinesAsSegments(t *testing.T) {
	c := New(WithMinWords(2))
	if _, err := c.ReadText("alpha beta gamma delta\none two three four"); err != nil {
		t.Fatalf("ReadText: %v", err)
	}
	if _, ok := c.StateDictionary()["gamma"+prefixSeparator+"delta"]; !ok {
		t.Fatal("expected the first line's trailing prefix to be present")
	}
	for _, suffix := range c.StateDictionary()["gamma"+prefixSeparator+"delta"] {
		if suffix != endMarker {
			t.Fatalf("line 1 chained into line 2 via suffix %q", suffix)
		}
	}
}

func TestWalkStopsAtSegmentEnd(t *testing.T) {
	// No terminal punctuation anywhere: only the end marker can stop this walk
	// before maxWords.
	c := New(WithMinWords(2), WithMaxWords(500))
	got, err := c.ReadText("alpha beta gamma delta epsilon")
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}
	if got != "alpha beta gamma delta epsilon" {
		t.Fatalf("got %q, want the segment reproduced exactly once", got)
	}
}

func TestEndMarkerNeverReachesOutput(t *testing.T) {
	c := New(WithMinWords(2))
	for range 100 {
		got, err := c.ReadText(corpus)
		if err != nil {
			t.Fatalf("ReadText: %v", err)
		}
		if strings.Contains(got, endMarker) || strings.Contains(got, prefixSeparator) {
			t.Fatalf("sentinel leaked into output: %q", got)
		}
		c.Reset()
	}
}
