package ytsmoviedir

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
)

type YtsMovieDirParser struct {
	parser.DefaultPriority
	tmdbService   parser.TmdbService
	distOpService parser.DiskOpService
}

func (p *YtsMovieDirParser) Parse(entry *dirinfo.Entry) error {
	// check entry is single media file
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	if entry.Type != dirinfo.DirEntry {
		return fmt.Errorf("entry is not a dir entry")
	}

	// try match with dir name pattern
	r, err := regexMatchDirName(entry)
	if err != nil {
		return fmt.Errorf("failed to regex match dir name: %v", err)
	}

	// classify files
	classified, err := classifyFile(entry)
	if err != nil {
		return fmt.Errorf("failed to classify file: %v", err)
	}

	// unknown file list should be zero len
	if len(classified[unknownType]) != 0 {
		return fmt.Errorf("unknown file list is not zero len")
	}

	// media file list should be len one
	if len(classified[mediaType]) != 1 {
		return fmt.Errorf("media file list is not len one")
	}
	mediaFile := classified[mediaType][0]
	if mediaFile.BytesNum <= mediaFileBytesSizeMin {
		return fmt.Errorf("media file size is too small")
	}

	// try find media file same name subtitle file
	subtitleMap := make(map[string][]*dirinfo.File)
	sameNameSubFile, err := tryFindSameNameSubtitleFile(mediaFile, classified[subtitleType])
	if err == nil {
		subtitleMap[""] = append(subtitleMap[""], sameNameSubFile)
	}

	// process subtitle file
	for _, aSubtitle := range classified[subtitleType] {
		if aSubtitle == sameNameSubFile {
			continue
		}
		onlyFileName := strings.TrimSuffix(aSubtitle.Name, aSubtitle.Ext)
		lang, err := getLanguageFromFileName(onlyFileName)
		if err != nil {
			return fmt.Errorf("failed to get language from file name: %v", err)
		}
		subtitleMap[lang] = append(subtitleMap[lang], aSubtitle)
	}

	// get tmdbid from pattern or by search name
	tmdbid, err := regexMatchTmdbid(entry)
	if err != nil {
		urlOption := map[string]string{
			"include_adult": "true",
			"year":          fmt.Sprintf("%d", r.year),
		}
		tmdbSearchMovies, err := p.tmdbService.GetSearchMovies(r.name, urlOption)
		if err != nil {
			return fmt.Errorf("failed to get search movies: %v", err)
		}
		if len(tmdbSearchMovies.Results) == 0 {
			// slog.Info(fmt.Sprintf("%s fail to search tmdb movie by name", templateName),
			// 	slog.String("name", r.name),
			// 	slog.Int("year", r.year),
			// )
			return fmt.Errorf("failed to match tmdbid")
		}
		if len(tmdbSearchMovies.Results) > 1 {
			// slog.Info(fmt.Sprintf("%s fail to unique search tmdb movie by name", templateName),
			// 	slog.String("name", r.name),
			// 	slog.Int("year", r.year),
			// 	slog.Int("candidates", len(tmdbSearchMovies.Results)),
			// )
			return fmt.Errorf("more than one movie matched")
		}
		tmdbid = int(tmdbSearchMovies.Results[0].ID)
	}

	// search tmdb by id
	tmdbMovieDetail, err := p.tmdbService.GetMovieDetails(tmdbid, defaultUrlOptions)
	if err != nil {
		return fmt.Errorf("failed to get movie details: %v", err)
	}
	tmdbid = int(tmdbMovieDetail.ID)

	// logging
	// slog.Info(fmt.Sprintf("%s parser succ", templateName),
	// 	slog.Int("tmdbid", int(tmdbid)),
	// 	slog.String("originalTitle", tmdbMovieDetail.OriginalTitle),
	// 	slog.String("airDate", tmdbMovieDetail.ReleaseDate),
	// )

	// perform disk operations
	err = p.distOpService.RenameSingleMovieFile(entry, mediaFile, tmdbMovieDetail, diskop.OnAirMovie)
	if err != nil {
		return fmt.Errorf("failed to rename single movie file: %v", err)
	}
	err = p.distOpService.RenameMovieSubtiles(entry, subtitleMap, tmdbMovieDetail, diskop.OnAirMovie)
	if err != nil {
		return fmt.Errorf("failed to rename movie subtitles: %v", err)
	}

	// perform disk operation delete dir
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
	tmdbidRegexp = regexp.MustCompile(`movie tmdbid-(?P<tmdbid>\d+)`)
)

func regexMatchTmdbid(entry *dirinfo.Entry) (int, error) {
	match := tmdbidRegexp.FindStringSubmatch(entry.MyDirPath)
	if len(match) != 2 {
		return 0, fmt.Errorf("failed to regex match tmdbid")
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert tmdbid: %v", err)
	}
	return n, nil
}

const (
	mediaFileBytesSizeMin = 1024 * 1024 * 100 // 100 MB
)

var (
	regexDirName = regexp.MustCompile(`^(?P<name>.*) \((?P<year>\d{4})\) .*\[\d+p\]`)
)

type rawInfo struct {
	name string
	year int
}

func regexMatchDirName(entry *dirinfo.Entry) (*rawInfo, error) {
	match := regexDirName.FindStringSubmatch(entry.MyDirPath)
	if len(match) != 3 {
		return nil, fmt.Errorf("failed to match dir name")
	}
	r := &rawInfo{
		name: match[1],
	}
	n, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, fmt.Errorf("failed to convert year: %v", err)
	}
	r.year = n
	return r, nil
}

type fileType int

const (
	mediaType fileType = iota
	subtitleType
	uselessType
	unknownType
)

func classifyFile(entry *dirinfo.Entry) (map[fileType][]*dirinfo.File, error) {
	ret := make(map[fileType][]*dirinfo.File, len(entry.FileList))
	for _, file := range entry.FileList {
		if ok := utils.IsMediaExt(file.Ext); ok {
			ret[mediaType] = append(ret[mediaType], file)
			continue
		}
		if ok := utils.IsSubtitleExt(file.Ext); ok {
			ret[subtitleType] = append(ret[subtitleType], file)
			continue
		}
		if ok := isUselessFile(file); ok {
			ret[uselessType] = append(ret[uselessType], file)
			continue
		}
		ret[unknownType] = append(ret[unknownType], file)
	}
	return ret, nil
}

var (
	uselessExt = map[string]struct{}{
		".txt": {},
		".jpg": {},
	}
)

func isUselessFile(file *dirinfo.File) bool {
	_, ok := uselessExt[file.Ext]
	return ok
}

func tryFindSameNameSubtitleFile(mediaFile *dirinfo.File, subtitleFileList []*dirinfo.File) (*dirinfo.File, error) {
	for _, aSuntileFile := range subtitleFileList {
		if strings.TrimSuffix(aSuntileFile.Name, aSuntileFile.Ext) == strings.TrimSuffix(mediaFile.Name, mediaFile.Ext) {
			return aSuntileFile, nil
		}
	}
	return nil, fmt.Errorf("failed to find same name subtitle file")
}
