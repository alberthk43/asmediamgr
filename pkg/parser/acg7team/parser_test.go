package acg7team

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser/fakes"
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
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

func TestNormalName(t *testing.T) {
	entry := &dirinfo.Entry{
		Type:       dirinfo.DirEntry,
		MyDirPath:  "[BDrip] Some Name S04 [7³ACG]",
		MotherPath: "",
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "[BDrip] Some Name S04 [7³ACG]/Some Name S04E01.mp4",
				Name:            "Some Name S04E01.mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "[BDrip] Some Name S04 [7³ACG]/Some Name S04E02.mp4",
				Name:            "Some Name S04E02.mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "[BDrip] Some Name S04 [7³ACG]/Some Name S04E03.mp4",
				Name:            "Some Name S04E03.mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 1000,
			},
			{
				RelPathToMother: "[BDrip] Some Name S04 [7³ACG]/NCOP.mp4",
				Name:            "NCOP.mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 100,
			},
			{
				RelPathToMother: "[BDrip] Some Name S04 [7³ACG]/NCED.mp4",
				Name:            "NCED.mp4",
				Ext:             ".mp4",
				BytesNum:        1024 * 1024 * 100,
			},
		},
	}
	parser := &Acg7TeamParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithNeedDelDir(),
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				4,
				1,
				diskop.OnAirTv,
			),
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[1],
				tvDetail,
				4,
				2,
				diskop.OnAirTv,
			),
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[2],
				tvDetail,
				4,
				3,
				diskop.OnAirTv,
			),
		),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
