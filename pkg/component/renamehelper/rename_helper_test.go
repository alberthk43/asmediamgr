package renamehelper

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"testing"
)

func TestTargetMovieFilePath(t *testing.T) {
	/// arrange
	movie := &common.MatchedMovie{
		MatchedCommon: common.MatchedCommon{
			OriginalTitle: "xxx movie",
			TmdbID:        123456,
			Year:          1900,
		},
	}
	/// action
	path, err := TargetMovieFilePath(movie, "./target", ".mp4")
	if err != nil {
		t.Fatal(err)
	}
	/// assert
	checkPath(t, path,
		"./target",
		"xxx movie (1900) [tmdbid-123456]",
		"xxx movie (1900).mp4")
}

func checkPath(t *testing.T, path renamer.Path, target, dir, file string) {
	t.Helper()
	if len(path) != 3 {
		t.Fatal(len(path))
	}
	if path[0] != "./target" {
		t.Fatalf("%s target:%s\n", path[0], target)
	}
	if path[1] != dir {
		t.Fatalf("%s dir:%s\n", path[1], dir)
	}
	if path[2] != file {
		t.Fatalf("%s file:%s\n", path[2], file)
	}
}

func TestReplaceSpecialCharacters(t *testing.T) {
	/// arrange
	original := `XXX: YYY`
	/// action
	replaced := ReplaceSpecialCharacters(original)
	/// asserrt
	expect := `XXX  YYY`
	if replaced != expect {
		t.Fatal(replaced, expect)
	}
}
