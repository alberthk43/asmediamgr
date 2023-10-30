package singlemoviefile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher/tests"
	"testing"
)

func TestSingleMovieFileMatchRegexp(t *testing.T) {
	/// arrange
	mockTMDBClient := &tests.MockTmdbService{
		ExpectedMovie: map[int64]*common.MatchedMovie{
			123456: {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle:    "XXX",
					OriginalLanguage: "en",
					TmdbID:           123456,
					Adult:            false,
					Year:             2010,
				},
			},
		},
	}
	mockRenamer := &tests.MockRenamer{
		Expected: []renamer.RenameRecord{
			{
				Old: renamer.Path{
					"./motherpath",
					"xxx movie tmdbid-123456.mp4",
				},
				New: renamer.Path{
					"./target",
					"XXX (2010) [tmdbid-123456]",
					"XXX (2010).mp4",
				},
			},
			{
				Old: renamer.Path{
					"./motherpath",
					"movie tmdbid-123456.mp4",
				},
				New: renamer.Path{
					"./target",
					"XXX (2010) [tmdbid-123456]",
					"XXX (2010).mp4",
				},
			},
		},
	}
	var tests = []struct {
		name          string
		info          common.Info
		ok            bool
		tmdbID        int64
		OriginalTitle string
	}{
		{"case1", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: "xxx movie tmdbid-123456", Ext: ".mp4"},
			}}, true, 123456, "xxx movie"},
		{"case2", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: "movie tmdbid-123456", Ext: ".mp4"},
			}}, false, 123456, "xxx movie"},
	}
	smfM, err := NewSingleMovieFileMatcher(mockTMDBClient, mockRenamer, "./target")
	if err != nil {
		t.Fatal(err)
	}
	/// action
	/// assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := smfM.Match(&tt.info)
			if tt.ok {
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal(ok)
				}
			} else {
				if err == nil {
					t.Fatal(ok)
				}
			}
		})
	}
}
