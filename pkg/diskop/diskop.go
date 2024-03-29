package diskop

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	oldPath := filepath.Join(entry.MotherPath, old.RelPathToMother)
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
	slog.Info("succ to rename single tv episode file", slog.String("old", oldPath), slog.String("new", newPath))
	return nil
}

func tvDirName(tvDetail *tmdb.TVDetails) (string, error) {
	year, err := parseYearFromAirDate(tvDetail.FirstAirDate)
	if err != nil {
		return "", err
	}
	originalName := replaceSpecialChars(tvDetail.OriginalName)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]", originalName, year, tvDetail.ID), nil
}

func parseYearFromAirDate(airDate string) (int, error) {
	t, err := time.Parse("2006-01-02", airDate)
	if err != nil {
		return 0, err
	}
	return t.Year(), nil
}

func tvSeasonDirName(tvDetail *tmdb.TVDetails, season int) string {
	return fmt.Sprintf("Season %d", season)
}

func tvEpFileName(old *dirinfo.File, tvDetail *tmdb.TVDetails, season int, episode int) string {
	return fmt.Sprintf("S%02dE%02d%s", season, episode, old.Ext)
}

func replaceSpecialChars(dirName string) string {
	dirName = strings.ReplaceAll(dirName, "\\", " ")
	dirName = strings.ReplaceAll(dirName, "/", " ")
	dirName = strings.ReplaceAll(dirName, ":", " ")
	dirName = strings.ReplaceAll(dirName, "*", " ")
	dirName = strings.ReplaceAll(dirName, "?", " ")
	dirName = strings.ReplaceAll(dirName, "<", " ")
	dirName = strings.ReplaceAll(dirName, ">", " ")
	dirName = strings.TrimSpace(dirName)
	return dirName
}

func (dop *DiskOpService) RenameSingleMovieFile(entry *dirinfo.Entry, old *dirinfo.File,
	movieDetail *tmdb.MovieDetails, destType DestType) error {
	oldPath := filepath.Join(entry.MotherPath, old.RelPathToMother)
	movieDir, err := movieDirName(movieDetail)
	if err != nil {
		return fmt.Errorf("failed to get movieDirName: %v", err)
	}
	movieFileName, err := movieFileName(old, movieDetail)
	if err != nil {
		return fmt.Errorf("failed to get movieFileName: %v", err)
	}
	destDir, ok := dop.destPathMap[destType]
	if !ok {
		return fmt.Errorf("no destPath for destType: %v", destType)
	}
	allDirPath := filepath.Join(destDir, movieDir)
	err = os.MkdirAll(allDirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create dir: %v", err)
	}
	newPath := filepath.Join(destDir, movieDir, movieFileName)
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
	slog.Info("succ to rename single movie file", slog.String("old", oldPath), slog.String("new", newPath))
	return nil
}

func movieDirName(movieDetail *tmdb.MovieDetails) (string, error) {
	year, err := parseYearFromAirDate(movieDetail.ReleaseDate)
	if err != nil {
		return "", err
	}
	originalTitle := replaceSpecialChars(movieDetail.OriginalTitle)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]", originalTitle, year, movieDetail.ID), nil
}

func movieFileName(old *dirinfo.File, movieDetail *tmdb.MovieDetails) (string, error) {
	year, err := parseYearFromAirDate(movieDetail.ReleaseDate)
	if err != nil {
		return "", err
	}
	originalTitle := replaceSpecialChars(movieDetail.OriginalTitle)
	return fmt.Sprintf("%s (%d)%s", originalTitle, year, old.Ext), nil
}

func (dop *DiskOpService) RenameMovieSubtiles(entry *dirinfo.Entry, filesMap map[string][]*dirinfo.File,
	movieDetail *tmdb.MovieDetails, destType DestType) error {
	var err error
	movieDirName, err := movieDirName(movieDetail)
	if err != nil {
		return fmt.Errorf("failed to get movieDirName: %v", err)
	}
	for lang, files := range filesMap {
		for _, file := range files {
			err = dop.renameOneMovieSubtile(entry, file, movieDetail, movieDirName, destType, lang)
		}
	}
	return err
}

func (dop *DiskOpService) renameOneMovieSubtile(entry *dirinfo.Entry, file *dirinfo.File, movieDetail *tmdb.MovieDetails,
	movieDirName string, destType DestType, language string) error {
	oldPath := filepath.Join(entry.MotherPath, file.RelPathToMother)
	subtitleFileName, err := movieSubtitleFileName(file, movieDetail, language)
	if err != nil {
		return fmt.Errorf("failed to get movieSubtitleFileName: %v", err)
	}
	destDir, ok := dop.destPathMap[destType]
	if !ok {
		return fmt.Errorf("no destPath for destType: %v", destType)
	}
	newPath := filepath.Join(destDir, movieDirName, subtitleFileName)
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
	slog.Info("succ to rename movie subtitle file", slog.String("old", oldPath), slog.String("new", newPath))
	return nil
}

func movieSubtitleFileName(old *dirinfo.File, movieDetail *tmdb.MovieDetails, language string) (string, error) {
	year, err := parseYearFromAirDate(movieDetail.ReleaseDate)
	if err != nil {
		return "", err
	}
	originalTitle := replaceSpecialChars(movieDetail.OriginalTitle)
	if language == "" {
		return fmt.Sprintf("%s (%d)%s", originalTitle, year, old.Ext), nil
	} else {
		return fmt.Sprintf("%s (%d).%s%s", originalTitle, year, language, old.Ext), nil
	}
}

func (dop *DiskOpService) DelDirEntry(entry *dirinfo.Entry) error {
	if entry.Type != dirinfo.DirEntry {
		return fmt.Errorf("try to del non-dir entry: %v", entry)
	}
	dirPath := filepath.Join(entry.MotherPath, entry.MyDirPath)
	return os.RemoveAll(dirPath)
}

func (dop *DiskOpService) RenameTvMusicFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails,
	name string, destType DestType) error {
	oldPath := filepath.Join(entry.MotherPath, old.RelPathToMother)
	tvDir, err := tvDirName(tvDetail)
	if err != nil {
		return fmt.Errorf("failed to get tvDirName: %v", err)
	}
	seasonDir := "Music"
	fileName := name
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
	slog.Info("succ to rename single tv episode file", slog.String("old", oldPath), slog.String("new", newPath))
	return nil
}
