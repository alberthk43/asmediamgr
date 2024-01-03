package tvepfile

import (
	"testing"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parsers/fakes"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var (
	tvDetail = &tmdb.TVDetails{
		FirstAirDate: "2020-05-07",
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
				Name:            "[ANi] Some Name - 08 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
				Ext:             ".mp4",
				BytesNum:        123456789,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				8,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExplictTmdbId(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[ANi] Name Do Not Matter Here - 08 [1080P][Baha][WEB-DL][AAC AVC][CHT] tv tmdbid-123456789.mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				1,
				8,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithChineaseSeasonInfo(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[ANi] Some Name 第三季 - 99 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				3,
				99,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithUselessAgeRestrict(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[ANi] Name Do Not Matter Here 第九季 [年齡限制版] - 56 [1080P][Baha][WEB-DL][AAC AVC][CHT] tv tmdbid-123456789.mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				9,
				56,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithUselessRegionRestrict(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[ANi] Name Do Not Matter Here 第九季 [年齡限制版]（僅限港澳台地區） - 56 [1080P][Baha][WEB-DL][AAC AVC][CHT] tv tmdbid-123456789.mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				9,
				56,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithUselessRegionRestrictAnotherOrder(t *testing.T) {
	entry := &dirinfo.Entry{
		Type: dirinfo.FileEntry,
		FileList: []*dirinfo.File{
			{
				RelPathToMother: "",
				Name:            "[ANi] Some Name [年齡限制版]（僅限港澳台地區）第九季 - 56 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
				Ext:             ".mp4",
				BytesNum:        888888,
			},
		},
	}
	parser := &TvEpParser{
		tmdbService: fakeTmdbService,
		distOpService: fakes.NewFakeDiskOpService(
			fakes.WithRenameSingleTvEpFile(
				entry,
				entry.FileList[0],
				tvDetail,
				9,
				56,
				diskop.OnAirTv,
			)),
	}
	err := parser.Parse(entry)
	if err != nil {
		t.Fatal(err)
	}
}
