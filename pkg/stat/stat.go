package stat

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/config"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/utils"
)

// const (
// 	cronConf = "*/5 * * * * *"
// )

var (
	reporter = &statReporter{}
)

func RegisterMovieChecker(c StatChecker) {
	reporter.movieCheckers = append(reporter.movieCheckers, c)
}

// StatOpts is a struct that holds options to create a new Stat to run stat task
type StatOpts struct {
	Logger log.Logger
	Config *config.Configuration
}

// Stat is a struct that holds infomation to run stat task
type Stat struct {
	logger   log.Logger
	Config   *config.Configuration
	statChan chan *StatInfo
}

// StatInfo is a struct that holds single entry stat info
type StatInfo struct {
	MsgType    msgType
	Entry      *dirinfo.Entry // raw entry
	Lasterr    error
	MediaTp    mediaType
	Tmdbid     int64
	TotalSize  int64
	MovieStat  *MovieStat
	TvshowStat *TvshowStat
}

func NewStat(opts *StatOpts) *Stat {
	if opts.Logger == nil {
		opts.Logger = log.NewNopLogger()
	}
	st := &Stat{
		logger:   opts.Logger,
		Config:   opts.Config,
		statChan: make(chan *StatInfo, 1),
	}
	return st
}

func RunStat(st *Stat) {
	level.Info(st.logger).Log("msg", "start stat task")
	go st.statTask()
	go st.reporterTask()
	select {}
}

type MovieStat struct {
	MovieFileNum    int
	SubtitleFileNum int
	ShortFileNum    int
	UnknownFileNum  int
}

type TvshowStat struct {
}

type msgType int

const (
	msgTypeStat msgType = iota
	msgTypeStart
	msgTypeEnd
)

type mediaType int

const (
	mediaTypeTv mediaType = iota
	mediaTypeMovie
)

func (st *Stat) statTask() {
	// log.Printf("lv=info component=stat msg=\"full start stat task\"\n")
	st.statChan <- &StatInfo{MsgType: msgTypeStart}
	conf := st.Config
	for i := range conf.StatDirs {
		// log.Printf("lv=info component=stat msg=\"start stat task\" dir=%s\n", conf.StatDirs[i].DirPath)
		err := st.statDir(&conf.StatDirs[i])
		if err != nil {
			// log.Printf("lv=error component=stat msg=\"%v\"\n", err)
		}
		// log.Printf("lv=info component=stat msg=\"end stat task\" dir=%s\n", conf.StatDirs[i].DirPath)
	}
	st.statChan <- &StatInfo{MsgType: msgTypeEnd}
	// log.Printf("lv=info component=stat msg=\"full end stat task\"\n")
}

func (st *Stat) statDir(statDir *config.StatDir) error {
	entries, err := dirinfo.ScanMotherDir(statDir.DirPath)
	if err != nil {
		return fmt.Errorf("failed to scan statDir %v err %v", statDir, err)
	}
	for _, entry := range entries {
		err = st.statEntry(statDir, entry)
		if err != nil {
			// log.Printf("lv=error component=stat statDir %v msg=\"%v\"\n", statDir, err)
			continue
		}
	}
	return nil
}

func (st *Stat) statEntry(statDir *config.StatDir, entry *dirinfo.Entry) error {
	// log.Printf("lv=info component=stat msg=\"stat entry\" entry=%+v\n", entry)
	if entry.Type == dirinfo.FileEntry {
		st.statChan <- &StatInfo{Entry: entry, Lasterr: fmt.Errorf("entry is a file")}
		return nil
	}
	switch statDir.MediaType {
	case config.MediaTypeTv:
		return st.statTvEntry(entry)
	case config.MediaTypeMovie:
		return st.statMovieEntry(entry)
	default:
		// default is allowed, do nothing
	}
	return nil
}

func (st *Stat) statTvEntry(entry *dirinfo.Entry) error {
	stInfo := &StatInfo{
		MsgType: msgTypeStat,
		Entry:   entry,
		MediaTp: mediaTypeTv,
	}
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		stInfo.Lasterr = fmt.Errorf("tvshow tmdbid is not recognized")
		st.statChan <- stInfo
		return nil
	}
	stInfo.Tmdbid = tmdbid
	tvshowStat := &TvshowStat{}
	stInfo.TvshowStat = tvshowStat
	st.statChan <- stInfo
	return nil
}

func (st *Stat) statMovieEntry(entry *dirinfo.Entry) error {
	stInfo := &StatInfo{
		MsgType: msgTypeStat,
		Entry:   entry,
		MediaTp: mediaTypeMovie,
	}
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		stInfo.Lasterr = fmt.Errorf("movie tmdbid is not recognized")
		st.statChan <- stInfo
		return nil
	}
	stInfo.Tmdbid = tmdbid
	movieStat := &MovieStat{}
	stInfo.MovieStat = movieStat
	for _, aFile := range entry.FileList {
		stInfo.TotalSize += aFile.BytesNum
		if utils.IsSubtitleExt(aFile.Ext) {
			movieStat.SubtitleFileNum++
		} else if utils.IsMediaExt(aFile.Ext) {
			if strings.Contains(aFile.RelPathToMother, `/shorts/`) {
				movieStat.ShortFileNum++
			} else {
				movieStat.MovieFileNum++
			}
		}
	}
	st.statChan <- stInfo
	return nil
}

var (
	movieDirNameRegexp = regexp.MustCompile(`^.* \(\d{4}\) \[tmdbid-(?P<tmdbid>\d+)\]$`)
)

func getTmdbidFromEntry(entry *dirinfo.Entry) int64 {
	groups := movieDirNameRegexp.FindStringSubmatch(entry.MyDirPath)
	if len(groups) != 2 {
		return 0
	}
	tmdbidStr := groups[1]
	n, err := strconv.ParseInt(tmdbidStr, 10, 63)
	if err != nil {
		return 0
	}
	return n
}

func (st *Stat) reporterTask() {
	// log.Printf("lv=debug component=stat msg=\"start reporter task\"\n")
	rpt := reporter
	for stInfo := range st.statChan {
		switch stInfo.MsgType {
		case msgTypeStart:
			rpt.reset(st, stInfo)
		case msgTypeStat:
			rpt.gather(st, stInfo)
		case msgTypeEnd:
			rpt.report(st)
		}
	}
}

type statReporter struct {
	movies         map[int64][]StatInfo
	tvshows        map[int64][]StatInfo
	movieCheckers  []StatChecker
	tvshowCheckers []StatChecker
}

type StatChecker interface {
	Check(tmdbid int64, stInfoSlice []StatInfo) error
}

func (r *statReporter) reset(st *Stat, stInfo *StatInfo) {
	level.Info(st.logger).Log("msg", "reset reporter")
	r.movies = make(map[int64][]StatInfo)
	r.tvshows = make(map[int64][]StatInfo)
}

func (r *statReporter) gather(st *Stat, stInfo *StatInfo) {
	level.Debug(st.logger).Log("msg", "reporter gather", "stat", st)
	if stInfo.Lasterr != nil {
		level.Error(st.logger).Log("msg", "stat has error", "err", stInfo.Lasterr)
		return
	}
	switch stInfo.MediaTp {
	case mediaTypeTv:
		// log.Printf("lv=info component=report mediaType=%d stat=%+v\n", stat.MediaTp, stat)
		stats := r.tvshows[stInfo.Tmdbid]
		stats = append(stats, *stInfo)
		r.tvshows[stInfo.Tmdbid] = stats
	case mediaTypeMovie:
		// log.Printf("lv=info component=report mediaType=%d stat=%+v\n", stat.MediaTp, stat)
		stats := r.movies[stInfo.Tmdbid]
		stats = append(stats, *stInfo)
		r.movies[stInfo.Tmdbid] = stats
	default:
		// log.Printf("lv=error component=report msg=\"unknown media type\", mediaType=%d stat=%+v\n", st.MediaTp, st)
	}
}

func (r *statReporter) report(st *Stat) {
	level.Info(st.logger).Log("msg", "reporter report start")
	for _, checker := range r.movieCheckers {
		for tmdbid, stats := range r.movies {
			err := checker.Check(tmdbid, stats)
			if err != nil {
				level.Warn(st.logger).Log("msg", "movie checker failed", "err", err)
			}
		}
	}
	for _, checker := range r.tvshowCheckers {
		for tmdbid, stats := range r.tvshows {
			err := checker.Check(tmdbid, stats)
			if err != nil {
				level.Error(st.logger).Log("msg", "tvshow checker failed", "err", err)
			}
		}
	}
	level.Info(st.logger).Log("msg", "reporter report end")
}

func printAndDie(msg string) {
	// log.Printf("lv=fatal component=stat msg=\"%s\"\n", msg)
	os.Exit(1)
}
