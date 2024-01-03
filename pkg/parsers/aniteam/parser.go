package tvepfile

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type TvEpParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *TvEpParser) Parse(entry *dirinfo.Entry) error {
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

	// TODO try match predefined patterns

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

	// process raw parsed name from filename, remove some useless info
	resultInfo, err := processRaw(rawInfo)
	if err != nil {
		return fmt.Errorf("failed to process raw: %v", err)
	}
	seasonnum = resultInfo.seasonnum
	name = resultInfo.name

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

type resultInfo struct {
	name      string
	seasonnum int
}

var (
	seasonnumRegexp = regexp.MustCompile(`^(?P<name>.*) 第(?P<seasonname>.*)季$`)
)

// processRaw for additional features
//  1. ignore "[年齡限制版]"
//  2. gather seasonnum from "第三季" like
//  3. ignore "（僅限港澳台地區）"
func processRaw(raw *rawInfo) (*resultInfo, error) {
	res := raw.raw
	seasonnum := int(1)
	res = strings.ReplaceAll(res, "[年齡限制版]", "")
	res = strings.Trim(res, " ")
	res = strings.ReplaceAll(res, "（僅限港澳台地區）", "")
	res = strings.Trim(res, " ")
	match := seasonnumRegexp.FindStringSubmatch(res)
	if len(match) == 3 {
		res = match[1]
		var ok bool
		if seasonnum, ok = chineseToValue(match[2]); !ok {
			return nil, fmt.Errorf("failed to convert seasonnum to int")
		}
	}
	return &resultInfo{
		name:      res,
		seasonnum: seasonnum,
	}, nil
}

func chineseToValue(chnStr string) (num int, ok bool) {
	switch chnStr {
	case "零":
		num = 0
	case "一":
		num = 1
	case "二":
		num = 2
	case "三":
		num = 3
	case "四":
		num = 4
	case "五":
		num = 5
	case "六":
		num = 6
	case "七":
		num = 7
	case "八":
		num = 8
	case "九":
		num = 9
	default:
		num = -1
	}
	return num, num >= 0
}

type rawInfo struct {
	raw   string
	epnum int
}

var (
	fileRegexp = regexp.MustCompile(`.*\[ANi\] (?P<raw>.*) - (?P<epnum>\d+) .*\.mp4$`)
)

func regexMatchFileName(file *dirinfo.File) (*rawInfo, error) {
	match := fileRegexp.FindStringSubmatch(file.Name)
	if len(match) != 3 {
		return nil, fmt.Errorf("failed to match regex")
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
