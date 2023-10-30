//go:build integration
// +build integration

package tmdbhttp

import (
	"os"
	"testing"
)

func TestSearchTVByName(t *testing.T) {
	/// arrange
	realHttpClient, err := NewTmdbHttpClient(os.Getenv("TMDB_READ_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	/// action
	data, err := SearchTVByName(realHttpClient, "Ahsoka")
	if err != nil {
		t.Fatal(err)
	}
	/// assert
	expectTV(t, data, &TMDBTVResult{
		ID:               114461,
		OriginalLanguage: "en",
		OriginalName:     "Ahsoka",
		Adult:            false,
		FirstAirDate:     "2023-08-22",
	})
}

func expectTV(t *testing.T, got, want *TMDBTVResult) {
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
