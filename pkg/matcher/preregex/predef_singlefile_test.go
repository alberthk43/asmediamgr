package preregex

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher/tests"
	"testing"
)

func TestLoadConfigs(t *testing.T) {
	/// arrange
	configPath := "./tests"
	// action
	matcher, err := NewPreRegexMatcher(
		configPath,
		&tests.MockTmdbService{},
		&tests.MockRenamer{},
		"./test_target_path",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(matcher.preDefRegexp) != 2 {
		t.Fatal(len(matcher.preDefRegexp))
	}
	checkLoaded(t, &subRegexMatcher{tmdbid: 123456, season: 2}, &matcher.preDefRegexp[0])
	checkLoaded(t, &subRegexMatcher{tmdbid: 123457, season: 1}, &matcher.preDefRegexp[1])
}

func checkLoaded(t *testing.T, expect, actual *subRegexMatcher) {
	t.Helper()
	if expect.tmdbid != actual.tmdbid {
		t.Fatalf("expect tmdbid %d, actual %d", expect.tmdbid, actual.tmdbid)
	}
	if expect.season != actual.season {
		t.Fatalf("expect season %d, actual %d", expect.season, actual.season)
	}
}

func TestMatch(t *testing.T) {
	/// arrange
	info := &common.Info{
		DirPath: "./src",
		Subs: []common.Single{
			{
				Paths: []string{},
				Name:  "[TeamName] QWERTY - 10",
				Ext:   ".mp4",
				IsDir: false,
				Size:  1024 * 1022 * 200,
			},
		},
	}
	m, err := NewPreRegexMatcher(
		"./tests/",
		&tests.MockTmdbService{
			ExpectedTv: map[int64]*common.MatchedTV{
				123456: {
					MatchedCommon: common.MatchedCommon{
						OriginalTitle:    "Original Title",
						OriginalLanguage: "Original Language",
						TmdbID:           123456,
						Adult:            false,
						Year:             2010,
					},
					Season: 2,
					EpNum:  10,
				},
			},
		},
		&tests.MockRenamer{
			Expected: []renamer.RenameRecord{
				{
					Old: renamer.Path{
						"./src",
						"[TeamName] QWERTY - 10.mp4",
					},
					New: renamer.Path{
						"./target",
						"Original Title (2010) [tmdbid-123456]",
						"Season 2",
						"S02E10.mp4",
					},
				},
			},
		},
		"./target",
	)
	if err != nil {
		t.Fatal(err)
	}
	/// action
	ok, err := m.Match(info)
	/// assert
	if !ok || err != nil {
		t.Fatal(err)
	}
}
