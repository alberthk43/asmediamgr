package moviefile

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser/fakes"
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var (
	movieDetail = &tmdb.MovieDetails{
		ReleaseDate: "2020-05-07",
		ID:          123456789,
	}
	fakeTmdbService = fakes.NewFakeTmdbService(
		fakes.WithMovieQueryMapping("Some Name", &tmdb.SearchMovies{
			Page:         1,
			TotalResults: 1,
			TotalPages:   1,
			SearchMoviesResults: &tmdb.SearchMoviesResults{
				Results: []struct {
					VoteCount        int64   `json:"vote_count"`
					ID               int64   `json:"id"`
					Video            bool    `json:"video"`
					VoteAverage      float32 `json:"vote_average"`
					Title            string  `json:"title"`
					Popularity       float32 `json:"popularity"`
					PosterPath       string  `json:"poster_path"`
					OriginalLanguage string  `json:"original_language"`
					OriginalTitle    string  `json:"original_title"`
					GenreIDs         []int64 `json:"genre_ids"`
					BackdropPath     string  `json:"backdrop_path"`
					Adult            bool    `json:"adult"`
					Overview         string  `json:"overview"`
					ReleaseDate      string  `json:"release_date"`
				}{
					{
						OriginalTitle: "Some Original Name",
						ID:            123456789,
					},
				},
			},
		}),
		fakes.WithMovieIdMapping(123456789, movieDetail),
	)
)

func TestWithTmdbId(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "Name Not Matter Here movie tmdbid-123456789.mp4",
				Ext:             ".mp4",
				BytesNum:        12345678999,
			},
		},
	}
	parser := &MovieFileParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleMovieFile(
				entry,
				entry.FileList[0],
				movieDetail,
				diskop.OnAirMovie,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalName(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "Some Name.mp4",
				Ext:             ".mp4",
				BytesNum:        12345678999,
			},
		},
	}
	parser := &MovieFileParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleMovieFile(
				entry,
				entry.FileList[0],
				movieDetail,
				diskop.OnAirMovie,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
