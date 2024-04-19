package stat

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/utils"
)

// StatOpts is a struct that holds options to create a new Stat to run stat task
type StatOpts struct {
	Logger             log.Logger
	Interval           time.Duration
	InitWait           time.Duration
	TvDirs             []string
	MovieDirs          []string
	LargeMovieSize     int64
	LargeTvEpisodeSize int64
}

// Stat is a struct that holds infomation to run stat task
type Stat struct {
	logger    log.Logger
	interval  time.Duration
	initWait  time.Duration
	tvDirs    []string
	movieDirs []string

	movieStats    map[int]*movieStat
	movieCheckers []movieChecker

	tvStats    map[int]*tvStat
	tvCheckers []tvChecker
}

const (
	defaultStatInterval = 6 * time.Hour
	defaultInitWait     = 10 * time.Second
)

func NewStat(opts *StatOpts) (*Stat, error) {
	if opts.Logger == nil {
		opts.Logger = log.NewNopLogger()
	}
	if opts.Interval == 0 {
		opts.Interval = defaultStatInterval
	}
	if opts.InitWait == 0 {
		opts.InitWait = defaultInitWait
	}
	if len(opts.TvDirs) == 0 && len(opts.MovieDirs) == 0 {
		return nil, fmt.Errorf("no tv or movie dirs")
	}
	st := &Stat{
		logger:    opts.Logger,
		interval:  opts.Interval,
		initWait:  opts.InitWait,
		tvDirs:    opts.TvDirs,
		movieDirs: opts.MovieDirs,
	}
	st.movieCheckers = append(st.movieCheckers, &multipleMovieChecker{})
	if opts.LargeMovieSize > 0 {
		st.movieCheckers = append(st.movieCheckers, &largetMovieChecker{sizeThreshold: opts.LargeMovieSize})
	}
	st.tvCheckers = append(st.tvCheckers, &multipleTvEpisodeChecker{})
	if opts.LargeTvEpisodeSize > 0 {
		st.tvCheckers = append(st.tvCheckers, &largeTvEpisodeChecker{sizeThreshold: opts.LargeTvEpisodeSize})
	}
	return st, nil
}

type StatErr interface {
	error
	toMarkdownContent() (string, error)
}

type movieChecker interface {
	check(*movieStat) StatErr
}

type movieStat struct {
	paths      []string
	tmdbid     int
	totalSize  int64
	movieFiles []*fileInfo
}

type tvChecker interface {
	check(*tvStat) StatErr
}

type tvStat struct {
	pathes       []string
	tmdbid       int
	totalSize    int64
	episodeFiles map[tvEpisodeKey][]*fileInfo
}

type tvEpisodeKey struct {
	season, episode int
}

type fileInfo struct {
	path string
	size int64
}

func (st *Stat) Run() error {
	if st.initWait > 0 {
		level.Info(st.logger).Log("msg", "wait for init", "dur", st.initWait)
		initWaitTicker := time.NewTicker(st.initWait)
		<-initWaitTicker.C
		initWaitTicker.Stop()
	}
	st.statTask()
	ticker := time.NewTicker(st.interval)
	for range ticker.C {
		st.statTask()
	}
	return nil
}

type MovieStat struct {
	MovieFileNum    int
	SubtitleFileNum int
	ShortFileNum    int
	UnknownFileNum  int
}

type TvshowStat struct {
}

func (st *Stat) statTask() {
	level.Info(st.logger).Log("msg", "start stat task", "movieDirs", len(st.movieDirs), "tvDirs", len(st.tvDirs))
	var err error
	st.clearStats()
	err = st.statMovieDirs()
	if err != nil {
		level.Error(st.logger).Log("msg", "stat movie dirs failed", "err", err)
		return
	}
	err = st.statTvDirs()
	if err != nil {
		level.Error(st.logger).Log("msg", "stat tv dirs failed", "err", err)
		return
	}
	err = st.runMovieCheckers()
	if err != nil {
		level.Error(st.logger).Log("msg", "run movie checkers failed", "err", err)
		return
	}
	err = st.runTvCheckers()
	if err != nil {
		level.Error(st.logger).Log("msg", "run tv checkers failed", "err", err)
		return
	}
	level.Info(st.logger).Log("msg", "stat task done")
}

var (
	movieMarkdownFile   = "movie_stat.md"
	movieMarkdownHeader = "# Movie Stat Report\n\n"
	tvMarkdownFile      = "tv_stat.md"
	tvMarkdownHeader    = "# TV Stat Report\n\n"
)

func (st *Stat) runMovieCheckers() error {
	mdFile, err := os.Create(movieMarkdownFile)
	if err != nil {
		return err
	}
	defer mdFile.Close()
	_, err = mdFile.WriteString(movieMarkdownHeader)
	if err != nil {
		return err
	}
	for _, checker := range st.movieCheckers {
		for _, mStat := range st.movieStats {
			statErr := checker.check(mStat)
			if statErr == nil {
				continue
			}
			content, err := statErr.toMarkdownContent()
			if err != nil {
				level.Error(st.logger).Log("msg", "movie check to markdown content failed", "err", err)
				continue
			}
			_, err = mdFile.WriteString(content)
			if err != nil {
				level.Error(st.logger).Log("msg", "write to markdown file failed", "err", err)
				continue
			}
		}
	}
	return nil
}

func (st *Stat) runTvCheckers() error {
	mdFile, err := os.Create(tvMarkdownFile)
	if err != nil {
		return err
	}
	defer mdFile.Close()
	_, err = mdFile.WriteString(tvMarkdownHeader)
	if err != nil {
		return err
	}
	for _, checker := range st.tvCheckers {
		for _, tvStat := range st.tvStats {
			statErr := checker.check(tvStat)
			if statErr == nil {
				continue
			}
			content, err := statErr.toMarkdownContent()
			if err != nil {
				level.Error(st.logger).Log("msg", "tv check to markdown content failed", "err", err)
				continue
			}
			_, err = mdFile.WriteString(content)
			if err != nil {
				level.Error(st.logger).Log("msg", "write to markdown file failed", "err", err)
				continue
			}
		}
	}
	return nil
}

func (st *Stat) clearStats() {
	st.movieStats = make(map[int]*movieStat)
	st.tvStats = make(map[int]*tvStat)
}

func (st *Stat) statMovieDirs() error {
	for _, movieDir := range st.movieDirs {
		level.Info(st.logger).Log("msg", "stat movie dir", "dir", movieDir)
		entries, err := dirinfo.ScanMotherDir(movieDir)
		if err != nil {
			return err
		}
		err = st.statMovieDir(entries)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *Stat) statMovieDir(entries []*dirinfo.Entry) error {
	for _, entry := range entries {
		err := st.statMovieEntry(entry)
		if err != nil {
			level.Error(st.logger).Log("msg", "stat movie entry failed", "err", err)
			continue
		}
	}
	return nil
}

func (st *Stat) statMovieEntry(entry *dirinfo.Entry) error {
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		level.Error(st.logger).Log("msg", "failed to get tmdbid from entry", "path", filepath.Join(entry.MotherPath, entry.MyDirPath))
		return fmt.Errorf("failed to get tmdbid from entry")
	}
	mStat, ok := st.movieStats[tmdbid]
	if !ok {
		mStat = &movieStat{tmdbid: tmdbid}
		st.movieStats[tmdbid] = mStat
	}
	mStat.paths = append(mStat.paths, entry.MyDirPath)
	mStat.totalSize += getTotalSizeFromEntry(entry)
	mStat.movieFiles = st.getTotalMovieFiles(entry)
	return nil
}

func (st *Stat) statTvDirs() error {
	for _, tvDir := range st.tvDirs {
		level.Info(st.logger).Log("msg", "stat tv dir", "dir", tvDir)
		entries, err := dirinfo.ScanMotherDir(tvDir)
		if err != nil {
			return err
		}
		err = st.statTvDir(entries)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *Stat) statTvDir(entrys []*dirinfo.Entry) error {
	for _, entry := range entrys {
		err := st.statTvEntry(entry)
		if err != nil {
			level.Error(st.logger).Log("msg", "stat tv entry failed", "err", err)
			continue
		}
	}
	return nil

}

func (st *Stat) statTvEntry(entry *dirinfo.Entry) error {
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		level.Error(st.logger).Log("msg", "failed to get tmdbid from entry", "path", filepath.Join(entry.MotherPath, entry.MyDirPath))
		return fmt.Errorf("failed to get tmdbid from entry")
	}
	tStat, ok := st.tvStats[tmdbid]
	if !ok {
		tStat = &tvStat{tmdbid: tmdbid}
		st.tvStats[tmdbid] = tStat
	}
	tStat.pathes = append(tStat.pathes, entry.MyDirPath)
	tStat.totalSize += getTotalSizeFromEntry(entry)
	tStat.episodeFiles = st.getTotalTvEpisodeFiles(entry)
	return nil
}

var (
	movieDirNameRegexp = regexp.MustCompile(`^.* \(\d{4}\) \[tmdbid-(?P<tmdbid>\d+)\]$`)
)

func getTmdbidFromEntry(entry *dirinfo.Entry) int {
	groups := movieDirNameRegexp.FindStringSubmatch(entry.MyDirPath)
	if len(groups) != 2 {
		return 0
	}
	tmdbidStr := groups[1]
	n, err := strconv.Atoi(tmdbidStr)
	if err != nil {
		return 0
	}
	return n
}

func getTotalSizeFromEntry(entry *dirinfo.Entry) int64 {
	var totalSize int64
	for _, aFile := range entry.FileList {
		totalSize += aFile.BytesNum
	}
	return totalSize
}

func (st *Stat) getTotalMovieFiles(entry *dirinfo.Entry) []*fileInfo {
	fileInfos := make([]*fileInfo, 0)
	for _, file := range entry.FileList {
		if !utils.IsMediaExt(file.Ext) {
			continue
		}
		segments := strings.Split(file.RelPathToMother, string(filepath.Separator))
		if len(segments) != 2 {
			continue
		}
		fileInfos = append(fileInfos, &fileInfo{
			path: filepath.Join(entry.MotherPath, file.RelPathToMother),
			size: file.BytesNum,
		})
	}
	return fileInfos
}

var (
	tvEpisodePattern = regexp.MustCompile(`^.*S(?P<season>\d+)E(?P<episode>\d+).*$`)
)

func (st *Stat) getTotalTvEpisodeFiles(entry *dirinfo.Entry) map[tvEpisodeKey][]*fileInfo {
	episodeFiles := make(map[tvEpisodeKey][]*fileInfo)
	for _, file := range entry.FileList {
		if !utils.IsMediaExt(file.Ext) {
			continue
		}
		segments := strings.Split(file.RelPathToMother, string(filepath.Separator))
		if len(segments) != 3 {
			continue
		}
		fileName := segments[2]
		fileName = strings.TrimSuffix(fileName, file.Ext)
		groups := tvEpisodePattern.FindStringSubmatch(fileName)
		if len(groups) == 0 {
			continue
		}
		season := -1
		episode := -1
		var err error
		for i, name := range tvEpisodePattern.SubexpNames() {
			switch name {
			case "season":
				season, err = strconv.Atoi(groups[i])
				if err != nil {
					continue
				}
			case "episode":
				episode, err = strconv.Atoi(groups[i])
				if err != nil {
					continue
				}
			}
		}
		if season < 0 || episode < 0 {
			level.Error(st.logger).Log("msg", "failed to get season or episode", "invalid file", file.RelPathToMother)
		}
		key := tvEpisodeKey{season: season, episode: episode}
		if _, ok := episodeFiles[key]; !ok {
			episodeFiles[key] = make([]*fileInfo, 0)
		}
		episodeFiles[key] = append(episodeFiles[key], &fileInfo{
			path: filepath.Join(entry.MotherPath, file.RelPathToMother),
			size: file.BytesNum,
		})
	}
	return episodeFiles
}
