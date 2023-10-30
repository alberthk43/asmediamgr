package aniepfile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher/tests"
	"testing"
)

func TestAniEpMatcher(t *testing.T) {
	/// arrange
	mockTmdbService := &tests.MockTmdbService{
		ExpectedNameTv: map[string]*common.MatchedTV{
			"XXX": {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle:    "XXX",
					OriginalLanguage: "en",
					TmdbID:           123456,
					Adult:            false,
					Year:             2010,
				},
				Season: 1,
				EpNum:  10,
			},
			"YYY": {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle:    "YYY",
					OriginalLanguage: "en",
					TmdbID:           123457,
					Adult:            false,
					Year:             2011,
				},
				Season: 1,
				EpNum:  10,
			},
			"ZZZ": {
				MatchedCommon: common.MatchedCommon{
					OriginalTitle:    "ZZZ",
					OriginalLanguage: "en",
					TmdbID:           123458,
					Adult:            true,
					Year:             2012,
				},
				Season: 1,
				EpNum:  10,
			},
		},
	}
	mockRenamer := &tests.MockRenamer{
		Expected: []renamer.RenameRecord{
			{
				Old: renamer.Path{
					"./motherpath",
					"[ANi] XXX - 10 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
				},
				New: renamer.Path{
					"./target",
					"XXX (2010) [tmdbid-123456]",
					"Season 1",
					"S01E10.mp4",
				},
			},
			{
				Old: renamer.Path{
					"./motherpath",
					"[ANi] YYY（僅限港澳台地區） - 10 [1080P][Bilibili][WEB-DL][AAC AVC][CHT CHS].mp4",
				},
				New: renamer.Path{
					"./target",
					"YYY (2011) [tmdbid-123457]",
					"Season 1",
					"S01E10.mp4",
				},
			},
			{
				Old: renamer.Path{
					"./motherpath",
					"[ANi] ZZZ 第五季 - 10 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
				},
				New: renamer.Path{
					"./target",
					"ZZZ (2012) [tmdbid-123458]",
					"Season 5",
					"S05E10.mp4",
				},
			},
		},
	}
	var tests = []struct {
		name   string
		info   common.Info
		ok     bool
		tmdbID int64
	}{
		{"case1", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: "[ANi] XXX - 10 [1080P][Baha][WEB-DL][AAC AVC][CHT]", Ext: ".mp4"},
			}}, true, 123456},
		{"case2", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: "[ANi] YYY（僅限港澳台地區） - 10 [1080P][Bilibili][WEB-DL][AAC AVC][CHT CHS]", Ext: ".mp4"},
			}}, true, 123457},
		{"case3", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: "[ANi] ZZZ 第五季 - 10 [1080P][Baha][WEB-DL][AAC AVC][CHT]", Ext: ".mp4"},
			}}, true, 123458},
	}
	m, err := NewAnimeEpisodeFileMatcher(mockTmdbService, mockRenamer, "./target")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/// action
			ok, err := m.Match(&tt.info)
			/// assert
			if tt.ok {
				if err != nil {
					t.Error(err)
				}
				if !ok {
					t.Error(ok)
				}
			} else {
				if err == nil {
					t.Error(ok)
				}
			}
		})
	}
}
