package gmteam

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
)

type GmTeamParser struct {
	c *Configuration
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *GmTeamParser) Parse(entry *dirinfo.Entry) error {
	// check entry is single media file
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	if entry.Type != dirinfo.FileEntry {
		return fmt.Errorf("entry is not a file entry")
	}
	if len(entry.FileList) != 1 {
		return fmt.Errorf("entry has more than one file")
	}
	file := entry.FileList[0]
	if file.Ext != ".mp4" {
		return fmt.Errorf("entry is not a mp4 file")
	}
	tmdbid := int(0)
	name := string("")
	epnum := int(-1)
	seasonnum := int(-1)

	// try match with tmdbid pattern
	tmpTmdbid, err := regexMatchTmdbid(file)
	if err == nil {
		tmdbid = tmpTmdbid
	}

	// regex match filename
	rawInfo, err := regexMatchFileName(file)
	if err != nil {
		return fmt.Errorf("failed to regex match file: %v", err)
	}
	epnum = rawInfo.epnum

	// try match predefined patterns
	isPreMatched := false
	for _, predefined := range p.c.Predefined {
		if rawInfo.raw == predefined.Name {
			tmdbid = predefined.TmdbId
			seasonnum = predefined.SeasonNum
			isPreMatched = true
		}
	}

	// process raw parsed name from filename, remove some useless info
	if !isPreMatched {
		resultInfo, err := processRaw(rawInfo)
		if err != nil {
			return fmt.Errorf("failed to process raw: %v", err)
		}
		seasonnum = resultInfo.seasonnum
		name = resultInfo.name
	}

	// try match explicited season and episode number, S01E02 like
	if n1, n2, err := regexMatchSeasonEp(file); err == nil {
		seasonnum = n1
		epnum = n2
	}

	// if tmdbid<=0, search by name first, then check results only one
	if tmdbid <= 0 {
		tvSearchResults, err := p.tmdbService.GetSearchTVShow(name, defaultUrlOptions)
		if err != nil {
			return fmt.Errorf("failed to get search tv show: %v", err)
		}
		if tvSearchResults.TotalResults != 1 {
			return fmt.Errorf("search result is not 1")
		}
		tmdbid = int(tvSearchResults.Results[0].ID)
	}
	tmdbTvDetail, err := p.tmdbService.GetTVDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to get tv details: %v", err)
	}
	tmdbid = int(tmdbTvDetail.ID)

	// logging
	slog.Info("aniteam parser succ",
		slog.Int("tmdbid", int(tmdbid)),
		slog.String("originalName", tmdbTvDetail.OriginalName),
		slog.String("airDate", tmdbTvDetail.FirstAirDate),
		slog.Int("seasonNum", seasonnum),
		slog.Int("epNum", epnum),
	)

	// perform disk operations
	err = p.distOpService.RenameSingleTvEpFile(entry, file, tmdbTvDetail, seasonnum, epnum, diskop.OnAirTv)
	if err != nil {
		return fmt.Errorf("failed to rename single tv ep file: %v", err)
	}

	return nil
}

var (
	tmdbidRegexp = regexp.MustCompile(`tv tmdbid-(?P<tmdbid>\d+)`)
)

func regexMatchTmdbid(file *dirinfo.File) (int, error) {
	match := tmdbidRegexp.FindStringSubmatch(file.Name)
	if len(match) != 2 {
		return 0, fmt.Errorf("failed to match regex")
	}
	n, err := strconv.ParseInt(match[1], 10, 31)
	if err != nil {
		return 0, fmt.Errorf("failed to convert tmdbid to int64")
	}
	return int(n), nil
}

type rawInfo struct {
	raw   string
	epnum int
}

var (
	fileRegexp            = regexp.MustCompile(`^\[GM-Team\]\[国漫\]\[(?P<name>.*)\]\[.*\]\[\d{4}\]\[(?P<epnum>\d+)\].*\.mp4$`)
	onlyEnglishNameRegexp = regexp.MustCompile(`^\[GM-Team\]\[国漫\]\[(?P<name>.*)\]\[\d{4}\]\[(?P<epnum>\d+)\].*\.mp4$`)
)

func regexMatchFileName(file *dirinfo.File) (*rawInfo, error) {
	match := fileRegexp.FindStringSubmatch(file.Name)
	if len(match) != 3 {
		match = onlyEnglishNameRegexp.FindStringSubmatch(file.Name)
		if len(match) != 3 {
			return nil, fmt.Errorf("failed to match regex")
		}
	}
	n, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, fmt.Errorf("failed to convert epnum to int")
	}
	ret := &rawInfo{
		raw:   match[1],
		epnum: n,
	}
	return ret, nil
}

var (
	seasonnumRegexp = regexp.MustCompile(`^(?P<name>.*) 第(?P<seasonname>\d+)季$`)
)

type resultInfo struct {
	name      string
	seasonnum int
}

// processRaw for additional features
//  1. gather seasonnum from "第5季" like
func processRaw(raw *rawInfo) (*resultInfo, error) {
	res := raw.raw
	seasonnum := int(1)
	match := seasonnumRegexp.FindStringSubmatch(res)
	if len(match) == 3 {
		res = match[1]
		n, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, fmt.Errorf("failed to convert seasonnum to int")
		}
		seasonnum = n
	}
	return &resultInfo{
		name:      res,
		seasonnum: seasonnum,
	}, nil
}

var (
	seasonEpNumRegexp = regexp.MustCompile(`S(?P<season>\d+)E(?P<episode>\d+)`)
)

func regexMatchSeasonEp(file *dirinfo.File) (int, int, error) {
	match := seasonEpNumRegexp.FindStringSubmatch(file.Name)
	if len(match) != 3 {
		return 0, 0, fmt.Errorf("failed to match regex")
	}
	season, err := strconv.ParseInt(match[1], 10, 31)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert season to int64")
	}
	episode, err := strconv.ParseInt(match[2], 10, 31)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert episode to int64")
	}
	return int(season), int(episode), nil
}
