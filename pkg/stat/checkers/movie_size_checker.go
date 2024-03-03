package checkers

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/stat"
	"fmt"
	"path/filepath"
)

type MovieSizeChecker struct{}

const (
	sizeLimit = 10 * 1024 * 1024 * 1024
)

func (c *MovieSizeChecker) Check(tmdbid int64, stats []stat.Stat) error {
	stat := stats[0]
	if stat.TotalSize > sizeLimit {
		return fmt.Errorf("movie size too large for tmdbid %d, size %s, path %s", tmdbid, dirinfo.SizeToStr(stat.TotalSize), filepath.ToSlash(filepath.Join(stat.Entry.MotherPath, stat.Entry.MyDirPath)))
	}
	return nil
}

func init() {
	checker := &MovieSizeChecker{}
	stat.RegisterMovieChecker(checker)
}
