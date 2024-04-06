package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type DiskServiceOpts struct {
	Logger         log.Logger
	DryRunModeOpen bool
}

type DiskService struct {
	logger     log.Logger
	dryRunMode bool
}

func NewDiskService(opts *DiskServiceOpts) (*DiskService, error) {
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
	}
	return &DiskService{logger: opts.Logger, dryRunMode: opts.DryRunModeOpen}, nil
}

type TvEpisodeRenameTask struct {
	OldPath      string
	NewMotherDir string
	OriginalName string
	Year         int
	Tmdbid       int
	Season       int
	Episode      int
}

type MovieRenameTask struct {
	OldPath      string
	NewMotherDir string
	OriginalName string
	Year         int
	Tmdbid       int
}

func BuildNewEpisodePath(tvEpTask *TvEpisodeRenameTask) (dir, path string, err error) {
	ext := filepath.Ext(tvEpTask.OldPath)
	seasonDir := BuildRelTvEpDirPath(tvEpTask.OriginalName, tvEpTask.Year, tvEpTask.Tmdbid, tvEpTask.Season)
	epFile := BuildRelTvEpPath(tvEpTask.OriginalName, tvEpTask.Year, tvEpTask.Tmdbid, tvEpTask.Season, tvEpTask.Episode, ext)
	return filepath.Join(tvEpTask.NewMotherDir, seasonDir), filepath.Join(tvEpTask.NewMotherDir, epFile), nil
}

func BuildRelTvEpDirPath(originalName string, year, tmdbid, season int) string {
	originalName = EscapeSpecialChars(originalName)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]/Season %d", originalName, year, tmdbid, season)
}

func BuildRelTvEpPath(originalName string, year, tmdbid, season, episode int, ext string) string {
	originalName = EscapeSpecialChars(originalName)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]/Season %d/S%02dE%02d%s", originalName, year, tmdbid, season, season, episode, ext)
}

func BuildNewMovieDir(movieTask *MovieRenameTask) (dir, path string, err error) {
	ext := filepath.Ext(movieTask.OldPath)
	movieDir := BuildRelMovieDirPath(movieTask.OriginalName, movieTask.Year, movieTask.Tmdbid)
	moviePath := BuildRelMovieFilePath(movieTask.OriginalName, movieTask.Year, movieTask.Tmdbid, ext)
	return filepath.Join(movieTask.NewMotherDir, movieDir), filepath.Join(movieTask.NewMotherDir, moviePath), nil
}

func BuildRelMovieDirPath(originalName string, year, tmdbid int) string {
	originalName = EscapeSpecialChars(originalName)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]", originalName, year, tmdbid)
}

func BuildRelMovieFilePath(originalName string, year, tmdbid int, ext string) string {
	originalName = EscapeSpecialChars(originalName)
	return fmt.Sprintf("%s (%d) [tmdbid-%d]/%s (%d)%s", originalName, year, tmdbid, originalName, year, ext)
}

func EscapeSpecialChars(path string) string {
	path = strings.ReplaceAll(path, "\\", " ")
	path = strings.ReplaceAll(path, "/", " ")
	path = strings.ReplaceAll(path, ":", " ")
	path = strings.ReplaceAll(path, "*", " ")
	path = strings.ReplaceAll(path, "?", " ")
	path = strings.ReplaceAll(path, "<", " ")
	path = strings.ReplaceAll(path, ">", " ")
	path = strings.TrimSpace(path)
	return path
}

func (d *DiskService) RenameTvEpisode(task *TvEpisodeRenameTask) error {
	oldFile, err := os.Open(task.OldPath)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
	defer oldFile.Close()
	motherDirStat, err := os.Stat(task.NewMotherDir)
	if err != nil {
		return fmt.Errorf("Stat() error = %v", err)
	}
	motherDirMode := motherDirStat.Mode()
	seasonDir, epFilePath, err := BuildNewEpisodePath(task)
	if err != nil {
		return fmt.Errorf("BuildNewEpisodePath() error = %v", err)
	}
	if !d.dryRunMode {
		err = os.MkdirAll(seasonDir, motherDirMode)
		if err != nil {
			return fmt.Errorf("MkdirAll() error = %v", err)
		}
		err = os.Rename(task.OldPath, epFilePath)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename tv episode", "old", task.OldPath, "new", epFilePath, "dryrun", d.dryRunMode)
	return nil
}

func (d *DiskService) RenameMovie(task *MovieRenameTask) error {
	oldFile, err := os.Open(task.OldPath)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
	defer oldFile.Close()
	motherDirStat, err := os.Stat(task.NewMotherDir)
	if err != nil {
		return fmt.Errorf("Stat() error = %v", err)
	}
	motherDirMode := motherDirStat.Mode()
	movieDir, movieFilePath, err := BuildNewMovieDir(task)
	if err != nil {
		return fmt.Errorf("BuildNewMovieDir() error = %v", err)
	}
	if !d.dryRunMode {
		err := os.MkdirAll(movieDir, motherDirMode)
		if err != nil {
			return fmt.Errorf("MkdirAll() error = %v", err)
		}
		err = os.Rename(task.OldPath, movieFilePath)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename movie", "old", task.OldPath, "new", movieFilePath, "dryrun", d.dryRunMode)
	return nil
}
