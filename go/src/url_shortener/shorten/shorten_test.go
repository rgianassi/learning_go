package shorten

import "testing"

func TestShortenURL(t *testing.T) {
	tests := []struct {
		url            string
		wantShortenURL string
	}{
		{"", "da39a3"},
		{"https:/github.com/develersrl/powersoft-hmi", "f63377"},
	}

	for _, test := range tests {
		actualShortenURL := shorten(test.url)
		if actualShortenURL != test.wantShortenURL {
			t.Errorf("Shorten URL was incorrect, got: %s, want: %s.", actualShortenURL, test.wantShortenURL)
		}
	}
}
