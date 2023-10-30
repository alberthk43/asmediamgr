//go:build integration
// +build integration

package tmdbhttp

import (
	"os"
	"testing"
)

func TestSearchTVByID(t *testing.T) {
	/// arrange
	realHttpClient, err := NewTmdbHttpClient(os.Getenv("TMDB_READ_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	/// action
	data, err := SearchTVByTmdbID(realHttpClient, 30984)
	if err != nil {
		t.Fatal(err)
	}
	/// assert
	expect(t, data, &TMDBTVResult{
		ID:               30984,
		OriginalLanguage: "ja",
		OriginalName:     "BLEACH",
		Adult:            false,
		FirstAirDate:     "2004-10-05",
	})
}

func expect(t *testing.T, got, want *TMDBTVResult) {
	t.Helper()
	if got == nil || want == nil {
		t.Fatal("got or want nil")
	}
	if got.ID != want.ID {
		t.Errorf("got %d want %d", got.ID, want.ID)
	}
	if got.Adult != want.Adult {
		t.Errorf("got %v want %v", got.Adult, want.Adult)
	}
	if got.OriginalLanguage != want.OriginalLanguage {
		t.Errorf("got %s want %s", got.OriginalLanguage, want.OriginalLanguage)
	}
	if got.OriginalName != want.OriginalName {
		t.Errorf("got %s want %s", got.OriginalName, want.OriginalName)
	}
	if got.FirstAirDate != want.FirstAirDate {
		t.Errorf("got %s want %s", got.FirstAirDate, want.FirstAirDate)
	}
}
