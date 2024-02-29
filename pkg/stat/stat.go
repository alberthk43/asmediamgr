package stat

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/robfig/cron/v3"

	"asmediamgr/pkg/config"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/utils"
)

const (
	cronConf = "0 0 4 * * *"
)

var (
	conf     *config.Configuration
	statChan = make(chan *stat, 1)
)

type stat struct {
	msgType   msgType
	entry     *dirinfo.Entry // raw entry
	lasterr   error
	mediaTp   mediaType
	tmdbid    int64
	totalSize int64
	movieStat *movieStat
}

type movieStat struct {
	movieFileNum    int
	subtitleFileNum int
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
	statChan <- &stat{msgType: msgTypeStart}
	for i := range conf.StatDirs {
		err := statDir(&conf.StatDirs[i])
		if err != nil {
			log.Printf("lv=error component=stat msg=\"%v\"\n", err)
		}
	}
	statChan <- &stat{msgType: msgTypeEnd}
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
	if entry.Type == dirinfo.FileEntry {
		statChan <- &stat{entry: entry, lasterr: fmt.Errorf("entry is a file")}
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
	// TODO albert
	return nil
}

func statMovieEntry(entry *dirinfo.Entry) error {
	stat := &stat{
		msgType: msgTypeStat,
		entry:   entry,
		mediaTp: mediaTypeMovie,
	}
	tmdbid := getTmdbidFromEntry(entry)
	if tmdbid <= 0 {
		stat.lasterr = fmt.Errorf("tmdbid is not recognized")
		statChan <- stat
		return nil
	}
	stat.tmdbid = tmdbid
	movieStat := &movieStat{}
	stat.movieStat = movieStat
	for _, aFile := range entry.FileList {
		stat.totalSize += aFile.BytesNum
		if utils.IsSubtitleExt(aFile.Ext) {
			movieStat.subtitleFileNum++
		} else if utils.IsMediaExt(aFile.Ext) {
			movieStat.movieFileNum++
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

func RunStat(c *config.Configuration) {
	if conf == nil {
		printAndDie("config is nil")
	}
	conf = c
	myCron := cron.New(cron.WithSeconds())
	_, err := myCron.AddFunc(cronConf, statTask)
	if err != nil {
		printAndDie(fmt.Sprintf("error adding cron job: %v", err))
	}
	myCron.Start()
	select {}
}

func printAndDie(msg string) {
	log.Printf("lv=fatal component=stat msg=\"%s\"\n", msg)
	os.Exit(1)
}
