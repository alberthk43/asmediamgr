package tvepfile

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser/fakes"
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var (
	emptyConfig           = &Configuration{}
	withPredefindedConfig = &Configuration{
		Predefined: []Predefined{
			{
				Name:   "Predefined Name",
				TmdbId: 123456789,
			},
		},
	}
)

var (
	tvDetail = &tmdb.TVDetails{
		FirstAirDate: "2020-05-07",
		ID:           123456789,
	}
	fakeTmdbService = fakes.NewFakeTmdbService(
		fakes.WithTvQueryMapping("Some Name", &tmdb.SearchTVShows{
			Page:         1,
			TotalResults: 1,
			TotalPages:   1,
			SearchTVShowsResults: &tmdb.SearchTVShowsResults{
				Results: []struct {
					OriginalName     string   `json:"original_name"`
					ID               int64    `json:"id"`
					Name             string   `json:"name"`
					VoteCount        int64    `json:"vote_count"`
					VoteAverage      float32  `json:"vote_average"`
					PosterPath       string   `json:"poster_path"`
					FirstAirDate     string   `json:"first_air_date"`
					Popularity       float32  `json:"popularity"`
					GenreIDs         []int64  `json:"genre_ids"`
					OriginalLanguage string   `json:"original_language"`
					BackdropPath     string   `json:"backdrop_path"`
					Overview         string   `json:"overview"`
					OriginCountry    []string `json:"origin_country"`
				}{
					{
						OriginalName: "Some Original Name",
						ID:           123456789,
					},
				},
			},
		}),
		fakes.WithTvIdMapping(123456789, tvDetail),
	)
)

func TestNormalSuccEntry(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "Some.Name.S01e02.mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &TvEpParser{
		c:           emptyConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				2,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalSuccEntryWithNoDotName(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "Some Name.S01E02.mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &TvEpParser{
		c:           emptyConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				2,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithTmdbId(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "NameDoNot.Matter.S01E02 tv tmdbid-123456789.mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &TvEpParser{
		c:           emptyConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				2,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithPredefined(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "Predefined Name.S01E02.mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		c:           withPredefindedConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				2,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
