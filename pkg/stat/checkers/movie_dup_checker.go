package checkers

import (
	"asmediamgr/pkg/stat"
	"fmt"
	"path/filepath"
)

type MovieDupChecker struct{}

func (c *MovieDupChecker) Check(tmdbid int64, stats []stat.Stat) error {
	if len(stats) > 1 {
		dirs := ""
		for _, s := range stats {
			dirs += fmt.Sprintf("\"%s\" ", filepath.ToSlash(filepath.Join(s.Entry.MotherPath, s.Entry.MyDirPath)))
		}
		return fmt.Errorf("found duplicate movie entry for tmdbid %d dirs %s", tmdbid, dirs)
	}
	return nil
}

func init() {
	checker := &MovieDupChecker{}
	stat.RegisterChecker(checker)
}
