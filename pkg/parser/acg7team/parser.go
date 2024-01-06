package acg7team

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Acg7TeamParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

func (p *Acg7TeamParser) Parse(entry *dirinfo.Entry) error {
	if entry.Type != dirinfo.DirEntry {
		return fmt.Errorf("entry is not a dir entry")
	}

	// try match with dir name pattern
	name, seasonNum, err := regexMatchDirName(entry)
	if err != nil {
		return fmt.Errorf("failed to regex match dir name: %v", err)
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

	// try tmdbid pattern
	tmdbid, err := regexMatchTmdbId(entry)
	if err != nil {
		searchRets, err := p.tmdbService.GetSearchTVShow(name, defaultUrlOptions)
		if err != nil {
			return fmt.Errorf("failed to get search tv show: %v", err)
		}
		if len(searchRets.Results) == 0 {
			return fmt.Errorf("failed to match tmdbid")
		}
		if len(searchRets.Results) > 1 {
			return fmt.Errorf("too many search results")
		}
		tmdbid = int(searchRets.Results[0].ID)
	}

	// check again
	if tmdbid <= 0 {
		return fmt.Errorf("failed to match tmdbid")
	}
	if seasonNum < 0 {
		return fmt.Errorf("failed to match seasonNum")
	}

	// get tv detail
	detail, err := p.tmdbService.GetTVDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to get tv detail: %v", err)
	}

	// episode file op
	for _, episodeFile := range classified[episodeFile] {
		err = p.distOpService.RenameSingleTvEpFile(entry, episodeFile.file, detail, episodeFile.seasonNum, episodeFile.epNum, diskop.OnAirTv)
		if err != nil {
			return fmt.Errorf("failed to rename single tv episode file: %v", err)
		}
	}

	// music file op
	if len(classified[musicOpFile]) == 1 {
		opFile := classified[musicOpFile][0]
		err = p.distOpService.RenameTvMusicFile(entry, opFile.file, detail, opFile.file.Name, diskop.OnAirTv)
		if err != nil {
			return fmt.Errorf("failed to rename tv music file: %v", err)
		}
	}

	// music file ep
	if len(classified[musicEdFile]) == 1 {
		edFile := classified[musicEdFile][0]
		err = p.distOpService.RenameTvMusicFile(entry, edFile.file, detail, edFile.file.Name, diskop.OnAirTv)
		if err != nil {
			return fmt.Errorf("failed to rename tv music file: %v", err)
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

var (
	regexTmdbId = regexp.MustCompile(`tv tmdbid-(?P<tmdbid>\d+)`)
)

func regexMatchTmdbId(entry *dirinfo.Entry) (int, error) {
	match := regexTmdbId.FindStringSubmatch(entry.MyDirPath)
	if match == nil {
		return -1, fmt.Errorf("failed to regex match tmdbid")
	}
	tmdbid, err := strconv.Atoi(match[1])
	if err != nil {
		return -1, fmt.Errorf("failed to convert tmdbid: %v", err)
	}
	return tmdbid, nil
}

type fileType int

const (
	episodeFile fileType = iota
	musicOpFile
	musicEdFile
	unknown
)

type fileInfo struct {
	file      *dirinfo.File
	seasonNum int
	epNum     int
	fType     fileType
}

func classifyFiles(entry *dirinfo.Entry) (map[fileType][]*fileInfo, error) {
	ret := make(map[fileType][]*fileInfo, len(entry.FileList))
	for _, file := range entry.FileList {
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
		if strings.TrimSuffix(file.Name, file.Ext) == "NCOP" {
			ret[musicOpFile] = append(ret[musicOpFile], &fileInfo{
				file:  file,
				fType: musicOpFile,
			})
			continue
		}
		if strings.TrimSuffix(file.Name, file.Ext) == "NCED" {
			ret[musicEdFile] = append(ret[musicEdFile], &fileInfo{
				file:  file,
				fType: musicEdFile,
			})
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
	regexDirName = regexp.MustCompile(`\[BDrip\] (?P<name>.*) S(?P<season>\d+) \[.*7³ACG\]`)
)

func regexMatchDirName(entry *dirinfo.Entry) (name string, seasonNum int, err error) {
	if entry.Type != dirinfo.DirEntry {
		return "", -1, fmt.Errorf("entry is not a dir entry")
	}
	match := regexDirName.FindStringSubmatch(entry.MyDirPath)
	if match == nil {
		return "", -1, fmt.Errorf("failed to regex match dir name")
	}
	name = match[1]
	seasonNum, err = strconv.Atoi(match[2])
	if err != nil {
		return "", -1, fmt.Errorf("failed to convert season num: %v", err)
	}
	return name, seasonNum, nil
}
