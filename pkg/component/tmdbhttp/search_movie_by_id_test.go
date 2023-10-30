//go:build integration
// +build integration

package tmdbhttp

import (
	"os"
	"testing"
)

func TestSearchMovieByID(t *testing.T) {
	/// arrange
	realHttpClient, err := NewTmdbHttpClient(os.Getenv("TMDB_READ_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	/// action
	data, err := SearchMovieByTmdbID(realHttpClient, 1058790)
	if err != nil {
		t.Fatal(err)
	}
	/// assert
	expect(t, data, &TMDBMovieResult{
		ID:               1058790,
		OriginalLanguage: "zh",
		OriginalTitle:    "奇迹的再现",
		Adult:            false,
		ReleaseDate:      "1985-01-01",
	})
}

func expect(t *testing.T, got, want *TMDBMovieResult) {
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
	if got.OriginalTitle != want.OriginalTitle {
		t.Errorf("got %s want %s", got.OriginalTitle, want.OriginalTitle)
	}
	if got.ReleaseDate != want.ReleaseDate {
		t.Errorf("got %s want %s", got.ReleaseDate, want.ReleaseDate)
	}
}
