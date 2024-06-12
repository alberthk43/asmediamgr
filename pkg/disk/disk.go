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

type TvSubtitleRenameTask struct {
	OldPath      string
	NewMotherDir string
	OriginalName string
	Year         int
	Tmdbid       int
	Season       int
	Episode      int
	Language     string
}

type MovieRenameTask struct {
	OldPath      string
	NewMotherDir string
	OriginalName string
	Year         int
	Tmdbid       int
}

type MovieSubtitleRenameTask struct {
	OldPath      string
	NewMotherDir string
	OriginalName string
	Year         int
	Tmdbid       int
	Language     string
}

type MoveToTrashTask struct {
	Path     string
	TrashDir string
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

func BuildNewTvSubtitlePath(tvSubTask *TvSubtitleRenameTask) (dir, path string, err error) {
	ext := filepath.Ext(tvSubTask.OldPath)
	seasonDir := BuildRelTvEpDirPath(tvSubTask.OriginalName, tvSubTask.Year, tvSubTask.Tmdbid, tvSubTask.Season)
	subtileFile := BuildRelTvSubtitlePath(tvSubTask.OriginalName, tvSubTask.Year, tvSubTask.Tmdbid, tvSubTask.Season, tvSubTask.Episode, tvSubTask.Language, ext)
	return filepath.Join(tvSubTask.NewMotherDir, seasonDir), filepath.Join(tvSubTask.NewMotherDir, subtileFile), nil
}

func BuildRelTvSubtitlePath(originalName string, year, tmdbid, season, episode int, lang, ext string) string {
	originalName = EscapeSpecialChars(originalName)
	if lang == "" {
		return fmt.Sprintf("%s (%d) [tmdbid-%d]/Season %d/S%02dE%02d%s", originalName, year, tmdbid, season, season, episode, ext)
	} else {
		return fmt.Sprintf("%s (%d) [tmdbid-%d]/Season %d/S%02dE%02d.%s%s", originalName, year, tmdbid, season, season, episode, lang, ext)
	}
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

func BuildNewMovieSubtitleDir(movieTask *MovieSubtitleRenameTask) (dir, path string, err error) {
	ext := filepath.Ext(movieTask.OldPath)
	movieDir := BuildRelMovieDirPath(movieTask.OriginalName, movieTask.Year, movieTask.Tmdbid)
	subtitlePath := BuildRelMovieSubtitleFilePath(movieTask.OriginalName, movieTask.Year, movieTask.Tmdbid, movieTask.Language, ext)
	return filepath.Join(movieTask.NewMotherDir, movieDir), filepath.Join(movieTask.NewMotherDir, subtitlePath), nil
}

func BuildRelMovieSubtitleFilePath(originalName string, year, tmdbid int, lang, ext string) string {
	originalName = EscapeSpecialChars(originalName)
	if lang == "" {
		return fmt.Sprintf("%s (%d) [tmdbid-%d]/%s (%d)%s", originalName, year, tmdbid, originalName, year, ext)
	} else {
		return fmt.Sprintf("%s (%d) [tmdbid-%d]/%s (%d).%s%s", originalName, year, tmdbid, originalName, year, lang, ext)
	}
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (d *DiskService) RenameTvEpisode(task *TvEpisodeRenameTask) error {
	_, err := os.Stat(task.OldPath)
	if err != nil {
		return fmt.Errorf("stat old file error = %v", err)
	}
	motherDirStat, err := os.Stat(task.NewMotherDir)
	if err != nil {
		return fmt.Errorf("stat new mother dir error = %v", err)
	}
	motherDirMode := motherDirStat.Mode()
	seasonDir, epFilePath, err := BuildNewEpisodePath(task)
	if err != nil {
		return fmt.Errorf("run BuildNewEpisodePath error = %v", err)
	}
	if !d.dryRunMode {
		err = os.MkdirAll(seasonDir, motherDirMode)
		if err != nil {
			return fmt.Errorf("run MkdirAll error = %v", err)
		}
		if fileExists(epFilePath) {
			return os.ErrExist
		}
		err = os.Rename(task.OldPath, epFilePath)
		if err != nil {
			return fmt.Errorf("run Rename error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename tv episode", "old", task.OldPath, "new", epFilePath, "dryrun", d.dryRunMode)
	return nil
}

func (d *DiskService) RenameTvSubtitle(task *TvSubtitleRenameTask) error {
	_, err := os.Stat(task.OldPath)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
	motherDirStat, err := os.Stat(task.NewMotherDir)
	if err != nil {
		return fmt.Errorf("Stat() error = %v", err)
	}
	motherDirMode := motherDirStat.Mode()
	seasonDir, subtitleFilePath, err := BuildNewTvSubtitlePath(task)
	if err != nil {
		return fmt.Errorf("BuildNewTvSubtitlePath() error = %v", err)
	}
	if !d.dryRunMode {
		err := os.MkdirAll(seasonDir, motherDirMode)
		if err != nil {
			return fmt.Errorf("MkdirAll() error = %v", err)
		}
		if fileExists(subtitleFilePath) {
			return os.ErrExist
		}
		err = os.Rename(task.OldPath, subtitleFilePath)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename tv subtitle", "old", task.OldPath, "new", subtitleFilePath, "dryrun", d.dryRunMode)
	return nil
}

func (d *DiskService) RenameMovie(task *MovieRenameTask) error {
	_, err := os.Stat(task.OldPath)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
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
		if fileExists(movieFilePath) {
			return os.ErrExist
		}
		err = os.Rename(task.OldPath, movieFilePath)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename movie", "old", task.OldPath, "new", movieFilePath, "dryrun", d.dryRunMode)
	return nil
}

func (d *DiskService) RenameMovieSubtitle(task *MovieSubtitleRenameTask) error {
	_, err := os.Stat(task.OldPath)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
	motherDirStat, err := os.Stat(task.NewMotherDir)
	if err != nil {
		return fmt.Errorf("Stat() error = %v", err)
	}
	motherDirMode := motherDirStat.Mode()
	movieDir, movieSubtitleFilePath, err := BuildNewMovieSubtitleDir(task)
	if err != nil {
		return fmt.Errorf("BuildNewMovieDir() error = %v", err)
	}
	if !d.dryRunMode {
		err := os.MkdirAll(movieDir, motherDirMode)
		if err != nil {
			return fmt.Errorf("MkdirAll() error = %v", err)
		}
		if fileExists(movieSubtitleFilePath) {
			return os.ErrExist
		}
		err = os.Rename(task.OldPath, movieSubtitleFilePath)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "rename movie subtitle", "old", task.OldPath, "new", movieSubtitleFilePath, "dryrun", d.dryRunMode)
	return nil
}

func (d *DiskService) MoveToTrash(task *MoveToTrashTask) error {
	_, err := os.Stat(task.TrashDir)
	if err != nil {
		return fmt.Errorf("Stat() error = %v", err)
	}
	_, err = os.Stat(task.Path)
	if err != nil {
		return fmt.Errorf("Open() error = %v", err)
	}
	oldBase := filepath.Base(task.Path)
	trashTarget := filepath.Join(task.TrashDir, oldBase)
	if !d.dryRunMode {
		if fileExists(trashTarget) {
			return os.ErrExist
		}
		err = os.Rename(task.Path, trashTarget)
		if err != nil {
			return fmt.Errorf("Rename() error = %v", err)
		}
	}
	level.Info(d.logger).Log("msg", "move to trash", "old", task.Path, "new", trashTarget, "dryrun", d.dryRunMode)
	return nil
}
