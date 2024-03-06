package checkers

import (
	"fmt"
	"path/filepath"

	"asmediamgr/pkg/stat"
)

type MovieDirDupChecker struct{}

func (c *MovieDirDupChecker) Check(tmdbid int64, stInfoSlice []stat.StatInfo) error {
	dirs := ""
	totalMovieFileNum := 0
	for _, s := range stInfoSlice {
		dirs += fmt.Sprintf("\"%s\" ", filepath.ToSlash(filepath.Join(s.Entry.MotherPath, s.Entry.MyDirPath)))
		totalMovieFileNum += s.MovieStat.MovieFileNum
	}
	if len(stInfoSlice) > 1 {
		return fmt.Errorf("found duplicate movie entry for tmdbid %d dirs %s", tmdbid, dirs)
	}
	if totalMovieFileNum > 1 {
		return fmt.Errorf("found duplicate movie files for tmdbid %d dirs %s", tmdbid, dirs)
	}
	return nil
}

func init() {
	checker := &MovieDirDupChecker{}
	stat.RegisterMovieChecker(checker)
}
