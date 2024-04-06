package moviedir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/disk"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/utils"
)

const (
	name = "moviedir"
)

func init() {
	parser.RegisterParser(name, &MovieDir{})
}

type Config struct {
	Patterns []*Pattern
}

type Pattern struct {
	DirPatternStr         string             `toml:"dir_pattern"`
	MediaPatternStr       string             `toml:"media_pattern"`
	MediaFileAtLeast      string             `toml:"media_file_at_least"`
	SubtitlePattern       []*SubtitlePattern `toml:"subtitle_pattern"`
	DirPattern            *regexp.Regexp
	MediaPattern          *regexp.Regexp
	MediaFileAtLeastBytes int64
}

type SubtitlePattern struct {
	Language   string `toml:"language"`
	PatternStr string `toml:"pattern"`
	Pattern    *regexp.Regexp
}

type MovieDir struct {
	logger   log.Logger
	patterns []*Pattern
}

func (p *MovieDir) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	p.logger = logger
	cfg := &Config{}
	_, err = toml.DecodeFile(cfgPath, cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	for _, pattern := range cfg.Patterns {
		pattern.DirPattern, err = regexp.Compile(pattern.DirPatternStr)
		if err != nil {
			return 0, err
		}
		pattern.MediaPattern, err = regexp.Compile(pattern.MediaPatternStr)
		if err != nil {
			return 0, err
		}
		pattern.MediaFileAtLeastBytes, err = utils.SizeStringToBytesNum(pattern.MediaFileAtLeast)
		if err != nil {
			return 0, err
		}
		for _, subtitlePattern := range pattern.SubtitlePattern {
			subtitlePattern.Pattern, err = regexp.Compile(subtitlePattern.PatternStr)
			if err != nil {
				return
			}
		}
	}
	p.patterns = cfg.Patterns
	return 0, nil
}

func (p *MovieDir) IsDefaultEnable() bool {
	return true
}

func (p *MovieDir) Parse(entry *dirinfo.Entry, opts *parser.ParserMgrRunOpts) (ok bool, err error) {
	if entry.Type != dirinfo.DirEntry {
		return false, nil
	}
	if len(entry.FileList) <= 0 {
		return false, fmt.Errorf("no files in dir, entry: %s", entry.Name())
	}
	movieTargetDir, ok := opts.MediaTypeDirs[common.MediaTypeMovie]
	if !ok {
		return false, fmt.Errorf("movie target dir not found, entry: %s", entry.Name())
	}
	trashDir, ok := opts.MediaTypeDirs[common.MediaTypeTrash]
	if !ok {
		return false, fmt.Errorf("trash dir not found, entry: %s", entry.Name())
	}
	info, err := p.parse(entry)
	if err != nil {
		return false, fmt.Errorf("failed to parse: %w, entry: %s", err, entry.Name())
	}
	if info == nil {
		return false, nil
	}
	level.Info(p.logger).Log("msg", "matched", "dir", entry.Name(), "name", info.name, "originalName", info.originalName, "year", info.year, "tmdbid", info.tmdbid, "subs", len(info.subtitleFiles))
	diskService := parser.GetDefaultDiskService()
	err = diskService.RenameMovie(&disk.MovieRenameTask{
		OldPath:      filepath.Join(entry.MotherPath, info.mediaFile.RelPathToMother),
		NewMotherDir: movieTargetDir,
		OriginalName: info.originalName,
		Year:         info.year,
		Tmdbid:       info.tmdbid,
	})
	if err != nil {
		return false, fmt.Errorf("failed to rename movie: %w, entry: %s", err, entry.Name())
	}
	for lang, subtitleFile := range info.subtitleFiles {
		err = diskService.RenameMovieSubtitle(&disk.MovieSubtitleRenameTask{
			OldPath:      filepath.Join(entry.MotherPath, subtitleFile.RelPathToMother),
			NewMotherDir: movieTargetDir,
			OriginalName: info.originalName,
			Year:         info.year,
			Tmdbid:       info.tmdbid,
			Language:     lang,
		})
		if err != nil {
			level.Warn(p.logger).Log("msg", "failed to rename subtitle", "lang", lang, "err", err, "entry", entry.Name())
			continue
		}
	}
	err = diskService.MoveToTrash(&disk.MoveToTrashTask{
		Path:     filepath.Join(entry.MotherPath, entry.Name()),
		TrashDir: trashDir,
	})
	if err != nil {
		level.Warn(p.logger).Log("msg", "failed to move to trash", "err", err, "entry", entry.Name())
	}
	return ok, nil
}

type movieInfo struct {
	name          string
	originalName  string
	year          int
	tmdbid        int
	mediaFile     *dirinfo.File
	subtitleFiles map[string]*dirinfo.File
}

func (p *MovieDir) parse(entry *dirinfo.Entry) (*movieInfo, error) {
	for _, pattern := range p.patterns {
		info, err := p.matchPattern(entry, pattern)
		if err != nil {
			return nil, err
		}
		if info != nil {
			return info, nil
		}
	}
	return nil, nil
}

var (
	defaultTmdbUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *MovieDir) matchPattern(entry *dirinfo.Entry, pattern *Pattern) (info *movieInfo, err error) {
	groups := pattern.DirPattern.FindStringSubmatch(entry.Name())
	if len(groups) <= 0 {
		return nil, nil
	}
	info = &movieInfo{}
	for i, name := range pattern.DirPattern.SubexpNames() {
		switch name {
		case "name":
			info.name = groups[i]
		case "year":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.year = n
		case "tmdbid":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.tmdbid = n
		}
	}
	var mediaFiles []*dirinfo.File
	subtitleFils := make(map[string]*dirinfo.File)
	for _, file := range entry.FileList {
		if utils.IsMediaExt(file.Ext) && utils.FileAtLeast(file, pattern.MediaFileAtLeastBytes) {
			mediaGroups := pattern.MediaPattern.FindStringSubmatch(file.RelPathToMother)
			if len(mediaGroups) > 0 {
				mediaFiles = append(mediaFiles, file)
				continue
			}
		}
		if utils.IsSubtitleExt(file.Ext) {
			for _, subtitlePattern := range pattern.SubtitlePattern {
				subtitleGroups := subtitlePattern.Pattern.FindStringSubmatch(file.RelPathToMother)
				if len(subtitleGroups) > 0 {
					if _, ok := subtitleFils[subtitlePattern.Language]; ok {
						return nil, fmt.Errorf("multiple subtitle files found, language: %s", subtitlePattern.Language)
					}
					subtitleFils[subtitlePattern.Language] = file
					break
				}
			}
		}
	}
	if len(mediaFiles) <= 0 {
		return nil, fmt.Errorf("no media files found")
	}
	if len(mediaFiles) > 1 {
		return nil, fmt.Errorf("multiple media files found")
	}
	info.mediaFile = mediaFiles[0]
	info.subtitleFiles = subtitleFils
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid <= 0 {
		searchOpts := defaultTmdbUrlOptions
		if info.year > 0 {
			searchOpts["year"] = strconv.Itoa(info.year)
		}
		results, err := tmdbService.GetSearchMovies(info.name, searchOpts)
		if err != nil {
			return nil, err
		}
		if results.TotalResults == 0 {
			return nil, fmt.Errorf("no movie found")
		}
		if results.TotalResults > 1 {
			var hits []string
			for i := 0; i < 3 && i < len(results.Results); i++ {
				hits = append(hits, fmt.Sprintf("%s-%d", results.Results[i].Title, results.Results[i].ID))
			}
			return nil, fmt.Errorf("multiple movies found: %v", hits)
		}
		info.tmdbid = int(results.Results[0].ID)
	}
	detail, err := tmdbService.GetMovieDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, err
	}
	info.originalName = detail.OriginalTitle
	dt, err := common.ParseTmdbDateStr(detail.ReleaseDate)
	if err != nil {
		return nil, err
	}
	info.year = dt.Year
	return info, nil
}
