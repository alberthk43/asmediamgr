package diskop

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
)

type DiskOpService struct {
	destPathMap map[DestType]string
}

func NewDiskOpService(destPathMap map[DestType]string) (*DiskOpService, error) {
	return &DiskOpService{destPathMap: destPathMap}, nil
}

type DestType int

const (
	OnAirTv DestType = iota
	OnAirMovie
)

// RenameSingleTvEpFile rename single tv episode file from old to new
func (dop *DiskOpService) RenameSingleTvEpFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails,
	season int, episode int, destType DestType) error {
	oldPath := filepath.Join(entry.MotherPath, old.RelPathToMother, old.Name)
	tvDir, err := tvDirName(tvDetail)
	if err != nil {
		return fmt.Errorf("failed to get tvDirName: %v", err)
	}
	seasonDir := tvSeasonDirName(tvDetail, season)
	fileName := tvEpFileName(old, tvDetail, season, episode)
	destDir, ok := dop.destPathMap[destType]
	if !ok {
		return fmt.Errorf("no destPath for destType: %v", destType)
	}
	allDirPath := filepath.Join(destDir, tvDir, seasonDir)
	err = os.MkdirAll(allDirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create dir: %v", err)
	}
	newPath := filepath.Join(destDir, tvDir, seasonDir, fileName)
	newFileStat, err := os.Stat(newPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat new file: %v", err)
		}
	} else {
		return fmt.Errorf("new file already exists: %v", newFileStat)
	}
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to create dir: %v", err)
	}
	slog.Info("succ to rename single tv episode file", slog.String("old", oldPath), slog.String("new", newPath), slog.String("dir", allDirPath))
	return nil
}

func tvDirName(tvDetail *tmdb.TVDetails) (string, error) {
	year, err := parseYearFromAirDate(tvDetail.FirstAirDate)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s (%d) [tmdbid-%d]", tvDetail.OriginalName, year, tvDetail.ID), nil
}

func parseYearFromAirDate(airDate string) (int, error) {
	t, err := time.Parse("2006-01-02", airDate)
	if err != nil {
		return 0, err
	}
	return t.Year(), nil
}

func tvSeasonDirName(tvDetail *tmdb.TVDetails, season int) string {
	return fmt.Sprintf("Season %02d", season)
}

func tvEpFileName(old *dirinfo.File, tvDetail *tmdb.TVDetails, season int, episode int) string {
	return fmt.Sprintf("S%02dE%02d%s", season, episode, old.Ext)
}
