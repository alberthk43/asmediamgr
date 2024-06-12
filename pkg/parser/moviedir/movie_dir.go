package moviedir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v2"

	"github.com/albert43/asmediamgr/pkg/common"
	"github.com/albert43/asmediamgr/pkg/dirinfo"
	"github.com/albert43/asmediamgr/pkg/disk"
	"github.com/albert43/asmediamgr/pkg/parser"
	"github.com/albert43/asmediamgr/pkg/utils"
)

const (
	name = "moviedir"
)

func init() {
	parser.RegisterParser(name, &MovieDir{})
}

type Config struct {
	Patterns []*Pattern `yaml:"patterns"`
}

type Pattern struct {
	Name                  string             `yaml:"name"`
	DirPattern            string             `yaml:"dir_pattern"`
	MediaPattern          *string            `yaml:"media_pattern"`
	SubtitlePattern       []*SubtitlePattern `yaml:"subtitle_pattern"`
	MediaFileAtLeast      string             `yaml:"media_file_at_least"`
	DirPatternRegexp      *regexp.Regexp
	MediaPatternRegexp    *regexp.Regexp
	MediaFileAtLeastBytes int64
}

type SubtitlePattern struct {
	Tag           string `yaml:"tag"`
	Pattern       string `yaml:"pattern"`
	PatternRegexp *regexp.Regexp
}

type MovieDir struct {
	logger   log.Logger
	patterns []*Pattern
}

func (p *MovieDir) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	p.logger = logger
	cfg := &Config{}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	for _, pattern := range cfg.Patterns {
		pattern.DirPatternRegexp, err = regexp.Compile(pattern.DirPattern)
		if err != nil {
			return 0, err
		}
		if pattern.MediaPattern != nil {
			pattern.MediaPatternRegexp, err = regexp.Compile(*pattern.MediaPattern)
			if err != nil {
				return 0, err
			}
		}
		pattern.MediaFileAtLeastBytes, err = utils.SizeStringToBytesNum(pattern.MediaFileAtLeast)
		if err != nil {
			return 0, err
		}
		for _, subtitlePattern := range pattern.SubtitlePattern {
			subtitlePattern.PatternRegexp, err = regexp.Compile(subtitlePattern.Pattern)
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

type movieInfo struct {
	name          string
	originalName  string
	year          int
	tmdbid        int
	mediaFile     *dirinfo.File
	subtitleFiles map[string]*dirinfo.File
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
	if info.mediaFile != nil {
		err = diskService.RenameMovie(&disk.MovieRenameTask{
			OldPath:      filepath.Join(entry.MotherPath, info.mediaFile.RelPathToMother),
			NewMotherDir: movieTargetDir,
			OriginalName: info.originalName,
			Year:         info.year,
			Tmdbid:       info.tmdbid,
		})
		if err != nil {
			if os.IsExist(err) {
				level.Warn(p.logger).Log("msg", "movie already existed", "entry", entry.Name(), "err", err)
			} else {
				return false, err
			}
		}
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
			if os.IsExist(err) {
				level.Warn(p.logger).Log("msg", "subtitle already existed", "lang", lang, "err", err, "entry", entry.Name())
			} else {
				return false, err
			}
		}
	}
	err = diskService.MoveToTrash(&disk.MoveToTrashTask{
		Path:     filepath.Join(entry.MotherPath, entry.Name()),
		TrashDir: trashDir,
	})
	if err != nil {
		level.Warn(p.logger).Log("msg", "failed to move to trash", "err", err, "entry", entry.Name())
	}
	return true, nil
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

func (p *MovieDir) matchPattern(entry *dirinfo.Entry, pattern *Pattern) (info *movieInfo, err error) {
	groups := pattern.DirPatternRegexp.FindStringSubmatch(entry.Name())
	if len(groups) <= 0 {
		return nil, nil
	}
	info = &movieInfo{}
	for i, name := range pattern.DirPatternRegexp.SubexpNames() {
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
	subtitleFilsMapping := make(map[string]*dirinfo.File) // subtitle tag -> file
	// try find media files, if no pattern, then see all matched
	for _, file := range entry.FileList {
		if !utils.IsMediaExt(file.Ext) || !utils.FileAtLeast(file, pattern.MediaFileAtLeastBytes) {
			continue
		}
		if pattern.MediaPattern == nil {
			mediaFiles = append(mediaFiles, file)
		} else {
			mediaGroups := pattern.MediaPatternRegexp.FindStringSubmatch(file.RelPathToMother)
			if len(mediaGroups) > 0 {
				continue
			}
		}
	}
	// only one movie media file is allowed, if not match, continue but only works with subtitles
	if len(mediaFiles) > 1 {
		return nil, fmt.Errorf("multiple media files found")
	}
	var allSubtitleFiles []*dirinfo.File
	for _, file := range entry.FileList {
		if utils.IsSubtitleExt(file.Ext) {
			allSubtitleFiles = append(allSubtitleFiles, file)
		}
	}
	var remainSubtitleFiles []*dirinfo.File
	for _, file := range allSubtitleFiles {
		found := false
		for _, subtitlePattern := range pattern.SubtitlePattern {
			subtitleGroups := subtitlePattern.PatternRegexp.FindStringSubmatch(file.RelPathToMother)
			if len(subtitleGroups) > 0 {
				if _, ok := subtitleFilsMapping[subtitlePattern.Tag]; !ok {
					subtitleFilsMapping[subtitlePattern.Tag] = file
					found = true
					break
				} else {
					return nil, fmt.Errorf("multiple subtitle files found for tag: %s", subtitlePattern.Tag)
				}
			}
		}
		if !found {
			remainSubtitleFiles = append(remainSubtitleFiles, file)
		}
	}
	if len(mediaFiles) == 1 { // if media file found, try found same name subtitle file, see as default subtitle
		info.mediaFile = mediaFiles[0]
		if _, ok := subtitleFilsMapping[""]; !ok {
			info.mediaFile = mediaFiles[0]
			for _, file := range remainSubtitleFiles {
				subtitleNameWithoutExt := file.Name[:len(file.Name)-len(file.Ext)-1]
				mediaNameWithoutExt := info.mediaFile.Name[:len(info.mediaFile.Name)-len(info.mediaFile.Ext)-1]
				if subtitleNameWithoutExt == mediaNameWithoutExt {
					subtitleFilsMapping[""] = file
				}
			}
		}
	} else { // if no media file found, then try to found one and only one subtitle file as default subtitle file
		if _, ok := subtitleFilsMapping[""]; !ok && len(remainSubtitleFiles) == 1 {
			subtitleFilsMapping[""] = remainSubtitleFiles[0]
		}
	}
	info.subtitleFiles = subtitleFilsMapping
	if info.mediaFile == nil && len(info.subtitleFiles) <= 0 {
		return info, nil
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid <= 0 {
		searchOpts := common.DefaultTmdbSearchOpts
		if info.year > 0 {
			searchOpts["year"] = strconv.Itoa(info.year)
		}
		results, err := tmdbService.GetSearchMovies(info.name, searchOpts)
		if err != nil {
			return nil, err
		}
		if results.TotalResults == 0 {
			return nil, fmt.Errorf("no movie found, name: %s, year: %d", info.name, info.year)
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
	detail, err := tmdbService.GetMovieDetails(info.tmdbid, common.DefaultTmdbSearchOpts)
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
