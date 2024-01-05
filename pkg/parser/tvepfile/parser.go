package tvepfile

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
)

type TvEpParser struct {
	c *Configuration
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

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
	if ok := utils.IsMediaExt(file.Ext); !ok {
		return fmt.Errorf("entry is not a media file")
	}

	tmdbid := int(0)
	epnum := int(-1)
	seasonnum := int(-1)

	// try match with filename pattern
	rawInfo, err := regexMatchFileName(file)
	if err != nil {
		return fmt.Errorf("failed to regex match file: %v", err)
	}
	name := rawInfo.name
	seasonnum = rawInfo.season
	epnum = rawInfo.ep

	// try match with tmdbid pattern
	tmdbid, err = regexMatchTmdbId(file)
	if err != nil {
		tmdbid, err = p.regexMatchPredefined(name)
		if err != nil {
			name := regulationName(rawInfo.name)
			tvSearchResults, err := p.tmdbService.GetSearchTVShow(name, defaultUrlOptions)
			if err != nil {
				return fmt.Errorf("failed to get search tv show: %v", err)
			}
			if len(tvSearchResults.Results) == 0 {
				return fmt.Errorf("failed to match tmdbid")
			}
			if len(tvSearchResults.Results) > 1 {
				return fmt.Errorf("more than one tv show matched")
			}
			tmdbid = int(tvSearchResults.Results[0].ID)
		}
	}
	if tmdbid <= 0 {
		return fmt.Errorf("failed to match tmdbid")
	}

	// search tv detail
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
	fileNameRegexp = regexp.MustCompile(`^(?P<name>.*)\.(S|s)(?P<season>\d+)(E|e)(?P<ep>\d+)`)
)

type rawInfo struct {
	name   string
	season int
	ep     int
}

func regexMatchFileName(file *dirinfo.File) (*rawInfo, error) {
	match := fileNameRegexp.FindStringSubmatch(file.Name)
	if len(match) != 6 {
		return nil, fmt.Errorf("failed to regex match filename")
	}
	rawInfo := &rawInfo{
		name: match[1],
	}
	var err error
	rawInfo.season, err = strconv.Atoi(match[3])
	if err != nil {
		return nil, fmt.Errorf("failed to convert season number: %v", err)
	}
	rawInfo.ep, err = strconv.Atoi(match[5])
	if err != nil {
		return nil, fmt.Errorf("failed to convert episode number: %v", err)
	}
	return rawInfo, nil
}

var (
	tmdbidRegexp = regexp.MustCompile(`tv tmdbid-(\d+)`)
)

func regexMatchTmdbId(file *dirinfo.File) (int, error) {
	match := tmdbidRegexp.FindStringSubmatch(file.Name)
	if len(match) != 2 {
		return 0, fmt.Errorf("failed to regex match tmdbid")
	}
	tmdbid, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert tmdbid to int: %v", err)
	}
	return tmdbid, nil
}

func (p *TvEpParser) regexMatchPredefined(name string) (int, error) {
	for _, predefined := range p.c.Predefined {
		if predefined.Name == name {
			return predefined.TmdbId, nil
		}
	}
	return 0, fmt.Errorf("failed to match predefined")
}

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func regulationName(name string) string {
	return strings.ReplaceAll(name, ".", " ")
}
