package tvepfile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher/tests"
	"testing"
)

func TestTvEpFileMatcher(t *testing.T) {
	/// arrange
	mockTmdbService := &tests.MockTmdbService{
		ExpectedNameTv: map[string]*common.MatchedTV{
			"XXX": {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle: "XXX",
					Year:          2021,
					TmdbID:        123,
				},
			},
		},
	}
	mockRenamerService := &tests.MockRenamer{
		Expected: []renamer.RenameRecord{
			{
				Old: renamer.Path{
					"montherpath",
					"XXX.S01E02.mkv",
				},
				New: renamer.Path{
					"targetpath",
					"XXX (2021) [tmdbid-123]",
					"Season 1",
					"S01E02.mkv",
				},
			},
		},
	}
	mockTargetService := &tests.MockTargetService{
		TargetDirPath: "targetpath",
	}
	m, err := NewTvEpisodeFileMatcher(
		"./tests/",
		mockTmdbService,
		mockRenamerService,
		mockTargetService,
	)
	if err != nil {
		t.Fatal(err)
	}
	/// action
	info := &common.Info{
		DirPath: "montherpath",
		Subs: []common.Single{
			{
				Name: "XXX.S01E02",
				Ext:  ".mkv",
				Size: 2 * 1024 * 1024 * 1024,
			},
		},
	}
	_, err = m.Match(info)
	if err != nil {
		t.Fatal(err)

	}
}

func TestTvEpFileMatcherTmdbid(t *testing.T) {
	/// arrange
	mockTmdbService := &tests.MockTmdbService{
		ExpectedTv: map[int64]*common.MatchedTV{
			123: {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle: "XXX",
					Year:          2021,
					TmdbID:        123,
				},
			},
		},
	}
	mockRenamerService := &tests.MockRenamer{
		Expected: []renamer.RenameRecord{
			{
				Old: renamer.Path{
					"montherpath",
					"XXX.S01E02 tmdbid-123.mkv",
				},
				New: renamer.Path{
					"targetpath",
					"XXX (2021) [tmdbid-123]",
					"Season 1",
					"S01E02.mkv",
				},
			},
		},
	}
	mockTargetService := &tests.MockTargetService{
		TargetDirPath: "targetpath",
	}
	m, err := NewTvEpisodeFileMatcher(
		"./tests/",
		mockTmdbService,
		mockRenamerService,
		mockTargetService,
	)
	if err != nil {
		t.Fatal(err)
	}
	/// action
	info := &common.Info{
		DirPath: "montherpath",
		Subs: []common.Single{
			{
				Name: "XXX.S01E02 tmdbid-123",
				Ext:  ".mkv",
				Size: 2 * 1024 * 1024 * 1024,
			},
		},
	}
	_, err = m.Match(info)
	if err != nil {
		t.Fatal(err)

	}
}
