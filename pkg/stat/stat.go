package stat

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"asmediamgr/pkg/config"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/utils"
)

const (
	cronConf = "*/5 * * * * *"
)

var (
	conf     *config.Configuration
	statChan = make(chan *Stat, 1)
	reporter = &statReporter{}
)

func RegisterMovieChecker(c StatChecker) {
	reporter.movieCheckers = append(reporter.movieCheckers, c)
}

type Stat struct {
	MsgType    msgType
	Entry      *dirinfo.Entry // raw entry
	Lasterr    error
	MediaTp    mediaType
	Tmdbid     int64
	TotalSize  int64
	MovieStat  *MovieStat
	TvshowStat *TvshowStat
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

func statTask() {
	log.Printf("lv=info component=stat msg=\"full start stat task\"\n")
	statChan <- &Stat{MsgType: msgTypeStart}
	for i := range conf.StatDirs {
		log.Printf("lv=info component=stat msg=\"start stat task\" dir=%s\n", conf.StatDirs[i].DirPath)
		err := statDir(&conf.StatDirs[i])
		if err != nil {
			log.Printf("lv=error component=stat msg=\"%v\"\n", err)
		}
		log.Printf("lv=info component=stat msg=\"end stat task\" dir=%s\n", conf.StatDirs[i].DirPath)
	}
	statChan <- &Stat{MsgType: msgTypeEnd}
	log.Printf("lv=info component=stat msg=\"full end stat task\"\n")
}

func statDir(statDir *config.StatDir) error {
	entries, err := dirinfo.ScanMotherDir(statDir.DirPath)
	if err != nil {
		return fmt.Errorf("failed to scan statDir %v err %v", statDir, err)
	}
	for _, entry := range entries {
		err = statEntry(statDir, entry)
		if err != nil {
			log.Printf("lv=error component=stat statDir %v msg=\"%v\"\n", statDir, err)
			continue
		}
	}
	return nil
}

func statEntry(statDir *config.StatDir, entry *dirinfo.Entry) error {
	// log.Printf("lv=info component=stat msg=\"stat entry\" entry=%+v\n", entry)
	if entry.Type == dirinfo.FileEntry {
		statChan <- &Stat{Entry: entry, Lasterr: fmt.Errorf("entry is a file")}
		return nil
	}
	switch statDir.MediaType {
	case config.MediaTypeTv:
		return statTvEntry(entry)
	case config.MediaTypeMovie:
		return statMovieEntry(entry)
	default:
		// default is allowed, do nothing
	}
	return nil
}

func statTvEntry(entry *dirinfo.Entry) error {
	stat := &Stat{
		MsgType: msgTypeStat,
		Entry:   entry,
		MediaTp: mediaTypeTv,
	}
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		stat.Lasterr = fmt.Errorf("tvshow tmdbid is not recognized")
		statChan <- stat
		return nil
	}
	stat.Tmdbid = tmdbid
	tvshowStat := &TvshowStat{}
	stat.TvshowStat = tvshowStat
	statChan <- stat
	return nil
}

func statMovieEntry(entry *dirinfo.Entry) error {
	stat := &Stat{
		MsgType: msgTypeStat,
		Entry:   entry,
		MediaTp: mediaTypeMovie,
	}
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		stat.Lasterr = fmt.Errorf("movie tmdbid is not recognized")
		statChan <- stat
		return nil
	}
	stat.Tmdbid = tmdbid
	movieStat := &MovieStat{}
	stat.MovieStat = movieStat
	for _, aFile := range entry.FileList {
		stat.TotalSize += aFile.BytesNum
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
	statChan <- stat
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

func reporterTask() {
	log.Printf("lv=debug component=stat msg=\"start reporter task\"\n")
	rpt := reporter
	for stat := range statChan {
		switch stat.MsgType {
		case msgTypeStart:
			rpt.reset()
		case msgTypeStat:
			rpt.gather(stat)
		case msgTypeEnd:
			rpt.report()
		}
	}
}

type statReporter struct {
	movies         map[int64][]Stat
	tvshows        map[int64][]Stat
	movieCheckers  []StatChecker
	tvshowCheckers []StatChecker
}

type StatChecker interface {
	Check(tmdbid int64, st []Stat) error
}

func (r *statReporter) reset() {
	log.Printf("lv=info component=stat msg=\"reset reporter\"\n")
	r.movies = make(map[int64][]Stat)
	r.tvshows = make(map[int64][]Stat)
}

func (r *statReporter) gather(stat *Stat) {
	// log.Printf("lv=info component=report msg=\"reporter gathter\" stat=%+v\n", stat)
	if stat.Lasterr != nil {
		log.Printf("lv=error component=report msg=\"stat has error\" stat=%+v err=%v\n", stat, stat.Lasterr)
		return
	}
	switch stat.MediaTp {
	case mediaTypeTv:
		// log.Printf("lv=info component=report mediaType=%d stat=%+v\n", stat.MediaTp, stat)
		stats := r.tvshows[stat.Tmdbid]
		stats = append(stats, *stat)
		r.tvshows[stat.Tmdbid] = stats
	case mediaTypeMovie:
		// log.Printf("lv=info component=report mediaType=%d stat=%+v\n", stat.MediaTp, stat)
		stats := r.movies[stat.Tmdbid]
		stats = append(stats, *stat)
		r.movies[stat.Tmdbid] = stats
	default:
		log.Printf("lv=error component=report msg=\"unknown media type\", mediaType=%d stat=%+v\n", stat.MediaTp, stat)
	}
}

func (r *statReporter) report() {
	log.Printf("lv=info component=stat msg=\"reporter report start\"\n")
	for _, checker := range r.movieCheckers {
		for tmdbid, stats := range r.movies {
			err := checker.Check(tmdbid, stats)
			if err != nil {
				log.Printf("lv=error component=stat msg=\"movie checker failed\" err=%v\n", err)
			}
		}
	}
	for _, checker := range r.tvshowCheckers {
		for tmdbid, stats := range r.tvshows {
			err := checker.Check(tmdbid, stats)
			if err != nil {
				log.Printf("lv=error component=stat  msg=\"tvshow checker failed\" err=%v\n", err)
			}
		}
	}
	log.Printf("lv=info component=stat msg=\"reporter report end\"\n")
}

func RunStat(c *config.Configuration) {
	if c == nil {
		printAndDie("config is nil")
	}
	conf = c
	log.Printf("lv=info component=stat msg=\"starting stat\" spec=%s\n", cronConf)
	go statTask()
	// myCron := cron.New(cron.WithSeconds())
	// _, err := myCron.AddFunc(cronConf, statTask)
	// if err != nil {
	// 	printAndDie(fmt.Sprintf("error adding cron job: %v", err))
	// }
	// myCron.Start()
	go reporterTask()
	select {}
}

func printAndDie(msg string) {
	log.Printf("lv=fatal component=stat msg=\"%s\"\n", msg)
	os.Exit(1)
}
