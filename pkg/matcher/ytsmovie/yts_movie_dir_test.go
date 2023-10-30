package ytsmovie

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"fmt"
	"testing"
)

type mockTMDBClientWithTMDBID struct{}

func (*mockTMDBClientWithTMDBID) DoTmdbHTTP(url string, data interface{}) error {
	switch url {
	case "https://api.themoviedb.org/3/movie/123456":
		v, ok := data.(*tmdbhttp.TMDBMovieResult)
		if !ok {
			return fmt.Errorf("data type not match")
		}
		if url == "https://api.themoviedb.org/3/movie/123456" {
			v.ID = 123456
			v.OriginalTitle = "A Far Shore"
			v.OriginalLanguage = "en"
			v.Adult = false
			v.ReleaseDate = "2023-01-01"
			return nil
		}
	default:
		return fmt.Errorf("url not match")
	}
	return fmt.Errorf("url not match")
}

type mockRenamerWithTMDBID struct{}

func (*mockRenamerWithTMDBID) Rename(records []renamer.RenameRecord) error {
	var err error
	if len(records) != 1 {
		return fmt.Errorf("len not match:%d", len(records))
	}
	old := records[0].Old
	new := records[0].New
	if len(old) < 2 {
		return fmt.Errorf("old path len < 2")
	}
	var expectOld, expectNew renamer.Path
	old1 := old[1]
	switch old1 {
	case dir1NameWithTMDBID:
		expectOld = renamer.Path{
			"./motherpath",
			dir1NameWithTMDBID,
			file1Full,
		}
		expectNew = renamer.Path{
			"./target",
			"A Far Shore (2023) [tmdbid-123456]",
			"A Far Shore (2023).mp4",
		}
	default:
		return fmt.Errorf("old path not match")
	}
	err = checkPath(expectOld, old)
	if err != nil {
		return err
	}
	err = checkPath(expectNew, new)
	if err != nil {
		return err
	}
	return nil
}

func checkPath(expect, actual renamer.Path) error {
	if len(expect) != len(actual) {
		return fmt.Errorf("len not match, expect:%d, actual:%d", len(expect), len(actual))
	}
	for i, v := range expect {
		a := actual[i]
		if v != a {
			return fmt.Errorf("path not match, expect:%s, actual:%s", v, a)
		}
	}
	return nil
}

var dir1NameWithTMDBID = "A Far Shore (2022) [JAPANESE] [1080p] [WEBRip] [5.1] [YTS.MX] movie tmdbid-123456"
var file1Name = "A Far Shore (2022) [JAPANESE] [1080p] [WEBRip] [5.1] [YTS.MX]"
var file1Ext = ".mp4"
var file1Full = fmt.Sprintf("%s%s", file1Name, file1Ext)
var sub1Name = file1Name
var sub1Ext = ".srt"
var sub1Full = fmt.Sprintf("%s%s", sub1Name, sub1Ext)

func TestYTSMovieDirMatchWithTMDBID(t *testing.T) {
	/// arrange
	mockTMDBClient := &mockTMDBClientWithTMDBID{}
	mockRenamer := &mockRenamerWithTMDBID{}

	var tests = []struct {
		name   string
		info   common.Info
		ok     bool
		tmdbID int64
	}{
		{"withTMDBID", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: dir1NameWithTMDBID, Paths: []string{}, IsDir: true},
				{Name: file1Name, Ext: file1Ext, Paths: []string{dir1NameWithTMDBID}, IsDir: false, Size: 2 * 1024 * 1024},
			}}, true, 123456},
	}
	smfM, err := NewYTSMovieDirMatcher(mockTMDBClient, mockRenamer, "./target")
	if err != nil {
		t.Fatal(err)
	}
	/// action
	/// assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := smfM.match(&tt.info)
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

type mockRenamerWithSameNameSubtitle struct{}

func (*mockRenamerWithSameNameSubtitle) Rename(records []renamer.RenameRecord) error {
	var err error
	if len(records) != 2 {
		return fmt.Errorf("len not match:%d", len(records))
	}
	for _, v := range records {
		old := v.Old
		new := v.New
		if len(v.Old) != 3 {
			return fmt.Errorf("old path len != 3")
		}
		var expectOld, expectNew renamer.Path
		switch v.Old[2] {
		case "A Far Shore (2022) [JAPANESE] [1080p] [WEBRip] [5.1] [YTS.MX].mp4":
			expectOld = renamer.Path{
				"./motherpath",
				dir1NameWithTMDBID,
				file1Full,
			}
			expectNew = renamer.Path{
				"./target",
				"A Far Shore (2023) [tmdbid-123456]",
				"A Far Shore (2023).mp4",
			}
		case "A Far Shore (2022) [JAPANESE] [1080p] [WEBRip] [5.1] [YTS.MX].srt":
			expectOld = renamer.Path{
				"./motherpath",
				dir1NameWithTMDBID,
				sub1Full,
			}
			expectNew = renamer.Path{
				"./target",
				"A Far Shore (2023) [tmdbid-123456]",
				"A Far Shore (2023).srt",
			}
		default:
			return fmt.Errorf("old path not match")
		}
		err = checkPath(expectOld, old)
		if err != nil {
			return err
		}
		err = checkPath(expectNew, new)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestYTSMovieDirMatchWithSameNameSubtitle(t *testing.T) {
	/// arrange
	mockTMDBClient := &mockTMDBClientWithTMDBID{}
	mockRenamer := &mockRenamerWithSameNameSubtitle{}

	var tests = []struct {
		name   string
		info   common.Info
		ok     bool
		tmdbID int64
	}{
		{"withTMDBID", common.Info{
			DirPath: "./motherpath",
			Subs: []common.Single{
				{Name: dir1NameWithTMDBID, Paths: []string{}, IsDir: true},
				{Name: file1Name, Ext: file1Ext, Paths: []string{dir1NameWithTMDBID}, IsDir: false, Size: 2 * 1024 * 1024},
				{Name: sub1Name, Ext: sub1Ext, Paths: []string{dir1NameWithTMDBID}, IsDir: false, Size: 100 * 1024},
			}}, true, 123456},
	}
	smfM, err := NewYTSMovieDirMatcher(mockTMDBClient, mockRenamer, "./target")
	if err != nil {
		t.Fatal(err)
	}
	/// action
	/// assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := smfM.match(&tt.info)
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

var subNameChineseChi = "Simplified.chi"
var subNameEnglish = "English.chi"
var subNameChineseTran = "Tranditional.chi"
