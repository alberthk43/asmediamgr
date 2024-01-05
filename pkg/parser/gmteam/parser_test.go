package gmteam

import (
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser/fakes"
)

var (
	emptyConfig           = &Configuration{}
	withPredefindedConfig = &Configuration{
		Predefined: []Predefined{
			{
				Name:      "Predefined Name 第4季",
				TmdbId:    123456789,
				SeasonNum: 4,
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
		fakes.WithQueryMapping("Some Name", &tmdb.SearchTVShows{
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
		fakes.WithIdMapping(123456789, tvDetail),
	)
)

func TestNormalSuccEntry(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[GM-Team][国漫][Some Name][English Name][2023][06][AVC][GB][1080P].mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &GmTeamParser{
		c:           emptyConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				6,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithSeasonNum(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.FileEntry,
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[GM-Team][国漫][Some Name 第4季][English Name 4th Season][2023][07][AVC][GB][1080P].mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &GmTeamParser{
		c:           emptyConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				4,
				7,
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
				Name:            "[GM-Team][国漫][Predefined Name 第4季][English Name 4th Season][2023][07][AVC][GB][1080P].mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &GmTeamParser{
		c:           withPredefindedConfig,
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				4,
				7,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
