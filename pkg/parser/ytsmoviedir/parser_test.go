package ytsmoviedir

import (
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser/fakes"
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

func TestNormalName(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/ Some Name (2020).mp4",
				Name:            "Some Name (2020).mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
		},
	}
	parser := &YtsMovieDirParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
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

func TestWithTmdbId(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  " Name Not Matter Any More (2020) [1080p] [WEBRip] [5.1] [YTS.MX] movie tmdbid-123456789",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: " Name Not Matter Any More (2020) [1080p] [WEBRip] [5.1] [YTS.MX] movie tmdbid-123456789/  Name Not Matter Any More (2020).mp4",
				Name:            " Name Not Matter Any More (2020).mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
		},
	}
	parser := &YtsMovieDirParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
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

func TestHasUselessFiles(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Some Name (2020).mp4",
				Name:            "Some Name (2020).mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Useless text file.txt",
				Name:            "Useless text file.txt",
				Ext:             ".txt",
				BytesNum:        100,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Useless img file.jpg",
				Name:            "Useless img file.jpg",
				Ext:             ".jpg",
				BytesNum:        100,
			},
		},
	}
	parser := &YtsMovieDirParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
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

func TestWithMediaSameNameSubtitle(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/ Some Name (2020).mp4",
				Name:            "Some Name (2020).mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/ Some Name (2020).srt",
				Name:            "Some Name (2020).srt",
				Ext:             ".srt",
				BytesNum:        100,
			},
		},
	}
	parser := &YtsMovieDirParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
			fakes.WithRenameSingleMovieFile(
				entry,
				entry.FileList[0],
				movieDetail,
				diskop.OnAirMovie,
			),
			fakes.WithRenameMovieSubtile(
				entry,
				entry.FileList[1],
				movieDetail,
				diskop.OnAirMovie,
				"",
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithMediaMultiSubtitle(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Some Name (2020).mp4",
				Name:            "Some Name (2020).mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Some Name (2020).srt",
				Name:            "Some Name (2020).srt",
				Ext:             ".srt",
				BytesNum:        100,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Subs/fre.srt",
				Name:            "fre.srt",
				Ext:             ".srt",
				BytesNum:        100,
			},
			{
				RelPathToMother: "Some Name (2020) [1080p] [WEBRip] [5.1] [YTS.MX]/Subs/Simplified.chi.srt",
				Name:            "Simplified.chi.srt",
				Ext:             ".srt",
				BytesNum:        100,
			},
		},
	}
	parser := &YtsMovieDirParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
			fakes.WithRenameSingleMovieFile(
				entry,
				entry.FileList[0],
				movieDetail,
				diskop.OnAirMovie,
			),
			fakes.WithRenameMovieSubtile(
				entry,
				entry.FileList[1],
				movieDetail,
				diskop.OnAirMovie,
				"",
			),
			fakes.WithRenameMovieSubtile(
				entry,
				entry.FileList[2],
				movieDetail,
				diskop.OnAirMovie,
				"fr",
			),
			fakes.WithRenameMovieSubtile(
				entry,
				entry.FileList[3],
				movieDetail,
				diskop.OnAirMovie,
				"zh-Hans",
			),
		),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
