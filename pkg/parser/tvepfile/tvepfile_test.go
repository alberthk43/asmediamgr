package tvepfile

// import (
// 	"testing"

// 	tmdb "github.com/cyruzin/golang-tmdb"
// 	"github.com/go-kit/log"

// 	"github.com/alberthk43/asmediamgr/pkg/dirinfo"
// 	"github.com/alberthk43/asmediamgr/pkg/parser"
// 	"github.com/alberthk43/asmediamgr/pkg/parser/fakes"
// )

// func TestLoadConfi(t *testing.T) {
// 	cfg, err := loadConfigFile("./testdata/tvepfile.toml")
// 	if err != nil {
// 		t.Fatalf("loadConfigFile() error = %v", err)
// 	}
// 	if len(cfg.Patterns) != 1 {
// 		t.Fatalf("loadConfigFile() got = %v, want = 1", len(cfg.Patterns))
// 	}
// 	pattern := cfg.Patterns[0]
// 	if pattern.PatternStr != `.*` {
// 		t.Fatalf("loadConfigFile() got = %v, want = .*", pattern.PatternStr)
// 	}
// 	if pattern.Tmdbid != 123456789 {
// 		t.Fatalf("loadConfigFile() got = %v, want = 123456789", pattern.Tmdbid)
// 	}
// 	if pattern.Season != 2 {
// 		t.Fatalf("loadConfigFile() got = %v, want = 2", pattern.Season)
// 	}
// 	if len(pattern.OptNames) != 2 {
// 		t.Fatalf("loadConfigFile() got = %v, want = 1", len(pattern.OptNames))
// 	}
// 	if pattern.OptNames[1] != "future_feature" {
// 		t.Fatalf("loadConfigFile() got = %v, want = future_feature", pattern.OptNames[1])
// 	}
// }

// var (
// 	tvDetail = &tmdb.TVDetails{
// 		FirstAirDate: "2020-05-07",
// 		ID:           123456789,
// 		OriginalName: "Some Original Name",
// 	}
// 	fakeTmdbService = fakes.NewFakeTmdbService(
// 		fakes.WithTvQueryMapping("Search Name", &tmdb.SearchTVShows{
// 			Page:         1,
// 			TotalResults: 1,
// 			TotalPages:   1,
// 			SearchTVShowsResults: &tmdb.SearchTVShowsResults{
// 				Results: []struct {
// 					OriginalName     string   `json:"original_name"`
// 					ID               int64    `json:"id"`
// 					Name             string   `json:"name"`
// 					VoteCount        int64    `json:"vote_count"`
// 					VoteAverage      float32  `json:"vote_average"`
// 					PosterPath       string   `json:"poster_path"`
// 					FirstAirDate     string   `json:"first_air_date"`
// 					Popularity       float32  `json:"popularity"`
// 					GenreIDs         []int64  `json:"genre_ids"`
// 					OriginalLanguage string   `json:"original_language"`
// 					BackdropPath     string   `json:"backdrop_path"`
// 					Overview         string   `json:"overview"`
// 					OriginCountry    []string `json:"origin_country"`
// 				}{
// 					{
// 						OriginalName: "Some Original Name",
// 						ID:           123456789,
// 					},
// 				},
// 			},
// 		}),
// 		fakes.WithTvIdMapping(123456789, tvDetail),
// 	)
// )

// func init() {
// 	parser.RegisterTmdbService(fakeTmdbService)
// }

// func compareTvEpInfo(t *testing.T, got, want *tvEpInfo) {
// 	t.Helper()
// 	// name do not matter after matching
// 	if got.originalName != want.originalName || got.season != want.season || got.episode != want.episode || got.tmdbid != want.tmdbid || got.year != want.year {
// 		t.Fatalf("compareTvEpInfo() got = %v, want = %v", got, want)
// 	}
// }

// func initTvEpFile(t *testing.T, parser *TvEpFile) {
// 	t.Helper()
// 	_, err := parser.Init("", log.NewNopLogger())
// 	if err != nil {
// 		t.Fatalf("Init() error = %v", err)
// 	}
// }

// func TestPreTmdbidAndSeason(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "Name Do NOT matter ep02 matter tv tmdbid-123456789.mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `ep(?P<episode>\d+).* tv tmdbid-(?P<tmdbid>\d+)$`,
// 				Tmdbid:     123456789,
// 				Season:     2,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       2,
// 		episode:      2,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }

// func TestPreTmdbidAndScrapedSeason(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "Name Do NOT matter S03e04 matter tv tmdbid-123456789.mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `[Ss](?P<season>\d+)[Ee](?P<episode>\d+).* tv tmdbid-(?P<tmdbid>\d+)$`,
// 				Tmdbid:     123456789,
// 				Season:     -1,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       3,
// 		episode:      4,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }

// func TestSearchByNameAndScrapedSeasonAndEpisode(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "Search Name S05E06.mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `(?P<name>.*) [Ss](?P<season>\d+)[Ee](?P<episode>\d+).*`,
// 				Tmdbid:     123456789,
// 				Season:     -1,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       5,
// 		episode:      6,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }

// func TestSearchByNameAndPreSeasonAndScrapedEpisode(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "[Some Title] Search Name - 23 [1080P][Source][WEB-DL][AAC AVC][Language].mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `\[Some Title\] (?P<name>.*) - (?P<episode>\d+)`,
// 				Tmdbid:     123456789,
// 				Season:     1,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       1,
// 		episode:      23,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }

// func TestSearchByNameAndPreSeasonAndScrapedEpisodeWithChineseSeasonNameOpt(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "[Some Title] Search Name 第二季 Meaningless Season Name - 24 [1080P][Source][WEB-DL][AAC AVC][CHT].mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `\[Some Title\] (?P<chineaseseasonname>.*) - (?P<episode>\d+)`,
// 				Tmdbid:     123456789,
// 				Season:     -1,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       2,
// 		episode:      24,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }

// func TestSearchByNameAndPreSeasonAndScrapedEpisodeWithNumberSeasonNameOpt(t *testing.T) {
// 	entry := &dirinfo.Entry{
// 		Type:       dirinfo.FileEntry,
// 		MotherPath: "",
// 		FileList: []*dirinfo.File{
// 			{
// 				RelPathToMother: "",
// 				Name:            "[Some Title] Search Name 第3季 Meaningless Season Name - 24 [1080P][Source][WEB-DL][AAC AVC][CHT].mp4",
// 				Ext:             ".mp4",
// 				BytesNum:        123456789,
// 			},
// 		},
// 	}
// 	parser := &TvEpFile{
// 		patterns: []*PatternConfig{
// 			{
// 				PatternStr: `\[Some Title\] (?P<numberseasonname>.*) - (?P<episode>\d+)`,
// 				Tmdbid:     123456789,
// 				Season:     -1,
// 			},
// 		},
// 	}
// 	initTvEpFile(t, parser)
// 	info, err := parser.parse(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	compareTvEpInfo(t, info, &tvEpInfo{
// 		originalName: "Some Original Name",
// 		season:       3,
// 		episode:      24,
// 		tmdbid:       123456789,
// 		year:         2020,
// 	})
// }
