package main

import "testing"

func TestShortenURL(t *testing.T) {
	tests := []struct {
		url                string
		expectedShortenURL string
	}{
		{"", "da39a3"},
		{"https:/github.com/develersrl/powersoft-hmi", "f63377"},
	}

	for _, test := range tests {
		actualShortenURL := shorten(test.url)
		if actualShortenURL != test.expectedShortenURL {
			t.Errorf("Shorten URL was incorrect, got: %s, want: %s.", actualShortenURL, test.expectedShortenURL)
		}
	}
}
