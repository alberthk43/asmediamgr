package bthdtv

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
	"fmt"
	"regexp"
	"strconv"
)

type BtHdtvParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

func (p *BtHdtvParser) Parse(entry *dirinfo.Entry) error {
	if entry.Type != dirinfo.DirEntry {
		return fmt.Errorf("entry is not a dir entry")
	}
	tmdbid, err := p.regexMatchTvTmdbid(entry)
	if err != nil {
		info, err := p.regexMatchDirName(entry)
		if err != nil {
			return fmt.Errorf("failed to regex match dir name: %v", err)
		}
		options := defaultUrlOptions
		options["year"] = fmt.Sprintf("%d", info.year)
		shows, err := p.tmdbService.GetSearchTVShow(info.name, options)
		if err != nil {
			return fmt.Errorf("failed to get search tv show: %v", err)
		}
		if shows.TotalResults == 0 {
			return fmt.Errorf("failed to get search tv show: no result")
		}
		if shows.TotalResults > 1 {
			return fmt.Errorf("failed to get search tv show: more than one result")
		}
		tmdbid = int(shows.Results[0].ID)
	}
	detail, err := p.tmdbService.GetTVDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to get tv detail: %v", err)
	}

	// classify files
	classified, err := classifyFiles(entry)
	if err != nil {
		return fmt.Errorf("failed to classify file: %v", err)
	}

	// check unknown file list is zero len
	if len(classified[unknown]) != 0 {
		return fmt.Errorf("unknown file list is not zero len")
	}

	// episode file op
	for _, episodeFile := range classified[episodeFile] {
		err = p.distOpService.RenameSingleTvEpFile(entry, episodeFile.file, detail, episodeFile.seasonNum, episodeFile.epNum, diskop.OnAirTv)
		if err != nil {
			return fmt.Errorf("failed to rename single tv episode file: %v", err)
		}
	}

	// del dir
	err = p.distOpService.DelDirEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to del dir entry: %v", err)
	}

	return nil
}

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

type fileType int

const (
	episodeFile fileType = iota
	unknown
)

type fileInfo struct {
	file      *dirinfo.File
	seasonNum int
	epNum     int
	fType     fileType
}

const (
	adMediaBytesNum = 1024 * 1024 * 2 // 2 MB
)

func classifyFiles(entry *dirinfo.Entry) (map[fileType][]*fileInfo, error) {
	ret := make(map[fileType][]*fileInfo, len(entry.FileList))
	for _, file := range entry.FileList {
		if utils.IsTorrentFile(file.Ext) {
			continue
		}
		if !utils.IsMediaExt(file.Ext) {
			ret[unknown] = append(ret[unknown], &fileInfo{
				file:  file,
				fType: unknown,
			})
			continue
		}
		season, ep, err := regexMatchEpisodeFile(file)
		if err == nil {
			ret[episodeFile] = append(ret[episodeFile], &fileInfo{
				file:      file,
				seasonNum: season,
				epNum:     ep,
				fType:     episodeFile,
			})
			continue
		}
		if file.BytesNum < adMediaBytesNum {
			continue
		}
		ret[unknown] = append(ret[unknown], &fileInfo{
			file:  file,
			fType: unknown,
		})
	}
	return ret, nil
}

var (
	regexSeasonFile = regexp.MustCompile(`[S|s](?P<seasonnum>\d+)[E|e](?P<epnum>\d+)`)
)

func regexMatchEpisodeFile(file *dirinfo.File) (season int, ep int, err error) {
	match := regexSeasonFile.FindStringSubmatch(file.Name)
	if match == nil {
		return -1, -1, fmt.Errorf("failed to regex match episode file")
	}
	season, err = strconv.Atoi(match[1])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to convert season num: %v", err)
	}
	ep, err = strconv.Atoi(match[2])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to convert ep num: %v", err)
	}
	return season, ep, nil
}

var (
	tmdbidPattern       = regexp.MustCompile(`tv tmdbid-(\d+)$`)
	dirNamePatternSlice = []*regexp.Regexp{
		regexp.MustCompile(`BTHDTV\.com.*】(?P<name>.*)\[全(\d+)集\]`),
		regexp.MustCompile(`DDHDTV\.com.*】(?P<name>.*)\[全(\d+)集\]`),
	}
	yearPattern          = regexp.MustCompile(`\.(\d{4})\.`)
	seasonPattern        = regexp.MustCompile(`\.S(\d+)\.`)
	seasonChinesePattern = regexp.MustCompile(`(?P<newname>.*) 第(?P<seasonname>.*)季`)
)

func removeChineseSeasonInName(name string) (newName string) {
	newName = name
	groups := seasonChinesePattern.FindStringSubmatch(name)
	if len(groups) == 3 {
		newName = groups[1]
	}
	return
}

func (p *BtHdtvParser) regexMatchDirName(entry *dirinfo.Entry) (info *dirMatchInfo, err error) {
	for _, pattern := range dirNamePatternSlice {
		info = &dirMatchInfo{}
		groups := pattern.FindStringSubmatch(entry.MyDirPath)
		if len(groups) == 0 {
			continue
		}
		subPattern := pattern.SubexpNames()
		for i, patternName := range subPattern {
			switch patternName {
			case "name":
				info.name = groups[i]
			case "season":
				n, err := strconv.ParseInt(groups[i], 10, 31)
				if err != nil {
					return nil, fmt.Errorf("failed to convert season num: %v, %s", err, groups[i])
				}
				info.season = int(n)
			case "year":
				n, err := strconv.ParseInt(groups[i], 10, 31)
				if err != nil {
					return nil, fmt.Errorf("failed to convert year: %v, %s", err, groups[i])
				}
				info.year = int(n)
			}
		}
		if info.name != "" {
			break
		}
	}
	if info == nil {
		return nil, fmt.Errorf("failed to regex match dir name")
	}
	if info.name == "" {
		return nil, fmt.Errorf("failed to regex match dir name")
	}
	groups := yearPattern.FindStringSubmatch(entry.MyDirPath)
	if len(groups) != 2 {
		return nil, fmt.Errorf("failed to regex match year")
	}
	n, err := strconv.ParseInt(groups[1], 10, 31)
	if err != nil {
		return nil, fmt.Errorf("failed to convert year: %v", groups[1])
	}
	info.year = int(n)
	groups = seasonPattern.FindStringSubmatch(entry.MyDirPath)
	n, err = strconv.ParseInt(groups[1], 10, 31)
	if err != nil {
		return nil, fmt.Errorf("failed to convert season: %v", groups[1])
	}
	info.season = int(n)
	info.name = removeChineseSeasonInName(info.name)
	return info, nil
}

func (p *BtHdtvParser) regexMatchTvTmdbid(entry *dirinfo.Entry) (tmdbid int, err error) {
	groups := tmdbidPattern.FindStringSubmatch(entry.MyDirPath)
	if len(groups) != 2 {
		return 0, fmt.Errorf("failed to regex match tmdbid")
	}
	n, err := strconv.ParseInt(groups[1], 10, 63)
	if err != nil {
		return 0, fmt.Errorf("failed to convert tmdbid: %v", err)
	}
	tmdbid = int(n)
	return
}

type dirMatchInfo struct {
	name   string
	year   int
	season int
}
