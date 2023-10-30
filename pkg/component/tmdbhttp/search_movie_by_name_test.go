//go:build integration
// +build integration

package tmdbhttp

import (
	"os"
	"testing"
)

func TestSearchMovieByName(t *testing.T) {
	/// arrange
	realHttpClient, err := NewTmdbHttpClient(os.Getenv("TMDB_READ_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	/// action
	data, err := SearchMovieByName(realHttpClient, "The Shawshank Redemption", 1994)
	if err != nil {
		t.Fatal(err)
	}
	/// assert
	expectMovie(t, data, &TMDBMovieResult{
		ID:               278,
		OriginalLanguage: "en",
		OriginalTitle:    "The Shawshank Redemption",
		Adult:            false,
		ReleaseDate:      "1994-09-23",
	})
}

func expectMovie(t *testing.T, got, want *TMDBMovieResult) {
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
