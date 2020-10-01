package shorten

import "testing"

func TestShortenURL(t *testing.T) {
	tests := []struct {
		url            string
		wantShortenURL string
	}{
		{"", "da39a3e"},
		{"https:/github.com/develersrl/powersoft-hmi", "f63377d"},
	}

	for _, test := range tests {
		actualShortenURL := Shorten(test.url)
		if actualShortenURL != test.wantShortenURL {
			t.Errorf("Shorten URL was incorrect, got: %s, want: %s.", actualShortenURL, test.wantShortenURL)
		}
	}
}
