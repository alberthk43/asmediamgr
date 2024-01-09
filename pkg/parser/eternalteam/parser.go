package eternalteam

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
)

type EternalTeamParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
	c             *Configuration
}

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *EternalTeamParser) Parse(entry *dirinfo.Entry) error {
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
	epnum := int(-1)
	seasonnum := int(-1)
	name := ""

	// regex match filename
	r, err := regexMatchFileName(file)
	if err != nil {
		return fmt.Errorf("failed to regexMatchFileName: %v", err)
	}
	epnum = r.epnum
	name = r.name

	// try match predefined
	r, err = p.regexMatchPredefined(name)
	if err != nil {
		n, err := regexMatchTmdbid(file)
		if err != nil {
			searchResults, err := p.tmdbService.GetSearchTVShow(name, defaultUrlOptions)
			if err != nil {
				return fmt.Errorf("failed to GetSearchTVShow: %v", err)
			}
			if len(searchResults.Results) == 0 {
				return fmt.Errorf("no search results")
			}
			if len(searchResults.Results) > 1 {
				return fmt.Errorf("more than one search results")
			}
			tmdbid = int(searchResults.Results[0].ID)
			seasonnum = 1
		} else {
			tmdbid = n
			seasonnum = 1
		}
	} else {
		tmdbid = r.tmdbid
		seasonnum = r.seasonnum
	}

	// check
	if tmdbid <= 0 || epnum < 0 || seasonnum < 0 {
		return fmt.Errorf("invalid tmdbid or epnum or seasonnum, %d, %d, %d", tmdbid, epnum, seasonnum)
	}

	// tmdb get detail
	detail, err := p.tmdbService.GetTVDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to GetTVDetails: %v", err)
	}

	// logging
	slog.Info(fmt.Sprintf("%s parser succ", templateName),
		slog.Int("tmdbid", int(tmdbid)),
		slog.String("originalName", detail.OriginalName),
		slog.String("airDate", detail.FirstAirDate),
		slog.Int("seasonNum", seasonnum),
		slog.Int("epNum", epnum),
	)

	// perform disk operations
	err = p.distOpService.RenameSingleTvEpFile(entry, file, detail, seasonnum, epnum, diskop.OnAirTv)
	if err != nil {
		return fmt.Errorf("failed to rename single tv ep file: %v", err)
	}

	return nil
}

func (p *EternalTeamParser) regexMatchPredefined(name string) (*rawInfo, error) {
	for _, predefined := range p.c.Predefined {
		if predefined.Name == name {
			return &rawInfo{
				tmdbid:    predefined.TmdbId,
				seasonnum: predefined.SeasonNum,
				name:      predefined.Name,
				epnum:     -1,
			}, nil
		}
	}
	return nil, fmt.Errorf("failed to match predefined")
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
	tmdbid    int
	seasonnum int
	name      string
	epnum     int
}

var (
	regexFileName = regexp.MustCompile(`^\[Eternal\]\[(?P<name>.*)\]\[(?P<epnum>\d+)\]`)
)

func regexMatchFileName(file *dirinfo.File) (*rawInfo, error) {
	match := regexFileName.FindStringSubmatch(file.Name)
	if len(match) != 3 {
		return nil, fmt.Errorf("failed to match regex")
	}
	n, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, fmt.Errorf("failed to convert epnum to int")
	}
	ret := &rawInfo{
		name:  match[1],
		epnum: n,
	}
	return ret, nil
}
