package moviefile

import (
	"fmt"
	"regexp"
	"strconv"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
)

type MovieFileParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

var (
	tmdbidRegexp = regexp.MustCompile(`movie tmdbid-(?P<tmdbid>\d+)`)
)

var (
	defaultUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *MovieFileParser) Parse(entry *dirinfo.Entry) error {
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
	name := string("")

	// match tmdbid pattern
	name = file.Name[:len(file.Name)-len(file.Ext)]
	tmdbid, err := regexMatchTmdbid(file)
	if err != nil {
		movieSearchResults, err := p.tmdbService.GetSearchMovies(name, defaultUrlOptions)
		if err != nil {
			return fmt.Errorf("failed to search movie: %v", err)
		}
		if len(movieSearchResults.Results) <= 0 {
			return fmt.Errorf("failed to search movie: no result")
		}
		if len(movieSearchResults.Results) != 1 {
			return fmt.Errorf("failed to search movie: more than one result")
		}
		tmdbid = int(movieSearchResults.Results[0].ID)
	}
	if tmdbid <= 0 {
		return fmt.Errorf("failed to get tmdbid")
	}

	// search movie detail
	tmdbMovieDetail, err := p.tmdbService.GetMovieDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to get movie detail: %v", err)
	}
	tmdbid = int(tmdbMovieDetail.ID)

	// logging
	// slog.Info(fmt.Sprintf("%s parser succ", templateName),
	// 	slog.Int("tmdbid", int(tmdbid)),
	// 	slog.String("originalName", tmdbMovieDetail.OriginalTitle),
	// 	slog.String("airDate", tmdbMovieDetail.ReleaseDate),
	// )

	// perform disk operations
	err = p.distOpService.RenameSingleMovieFile(entry, file, tmdbMovieDetail, diskop.OnAirMovie)
	if err != nil {
		return fmt.Errorf("failed to rename single movie file: %v", err)
	}

	return nil
}

func regexMatchTmdbid(file *dirinfo.File) (int, error) {
	matches := tmdbidRegexp.FindStringSubmatch(file.Name)
	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to match tmdbid pattern")
	}
	tmdbid, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert tmdbid to int: %v", err)
	}
	return tmdbid, nil
}
