package regexparser

import (
	"regexp"
	"testing"
)

func TestParseTmdbID(t *testing.T) {
	/// arrange
	tests := []struct {
		tname        string
		regex        *regexp.Regexp
		expectTmdbId int64
	}{
		{
			tname:        "tmdbid-123",
			regex:        regexp.MustCompile(`tmdbid-(?P<tmdbid>\d+)$`),
			expectTmdbId: 123,
		},
		{
			tname:        "XXX tmdbid-123",
			regex:        regexp.MustCompile(`tmdbid-(?P<tmdbid>\d+)$`),
			expectTmdbId: 123,
		},
		{
			tname:        "XXX tmdbid-123",
			regex:        regexp.MustCompile(`^(?P<name>.*) tmdbid-(?P<tmdbid>\d+)$`),
			expectTmdbId: 123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			tmdbID, err := ParseTmdbID(tt.regex, tt.tname)
			if err != nil {
				t.Error(err)
			}
			if tmdbID != tt.expectTmdbId {
				t.Errorf("expectTmdbID:%d, but:%d\n", tt.expectTmdbId, tmdbID)
			}
		})
	}
}
