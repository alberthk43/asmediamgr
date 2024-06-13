package tvdir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v2"

	"github.com/alberthk43/asmediamgr/pkg/common"
	"github.com/alberthk43/asmediamgr/pkg/dirinfo"
	"github.com/alberthk43/asmediamgr/pkg/disk"
	"github.com/alberthk43/asmediamgr/pkg/parser"
	"github.com/alberthk43/asmediamgr/pkg/utils"
)

const (
	name = "tvdir"
)

func init() {
	parser.RegisterParser(name, &TvDir{})
}

type Pattern struct {
	Name                    string `yaml:"name"`
	DirPattern              string `yaml:"dir_pattern"`
	EpisodePattern          string `yaml:"episode_pattern"`
	SubtitlePattern         string `yaml:"subtitle_pattern"`
	EpisodeFileAtLeast      string `yaml:"episode_file_at_least"`
	Tmdbid                  *int   `yaml:"tmdbid"`
	Season                  *int   `yaml:"season"`
	DirPatternRegexp        *regexp.Regexp
	EpisodePatternRegexp    *regexp.Regexp
	EpisodeFileAtLeastBytes int64
	SubtitlePatternRegexp   *regexp.Regexp
}

type Config struct {
	Patterns []*Pattern `yaml:"patterns"`
}

type TvDir struct {
	logger   log.Logger
	patterns []*Pattern
}

func (p *TvDir) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	p.logger = logger
	cfg := &Config{}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
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
		if pattern.DirPattern == "" {
			return 0, fmt.Errorf("no dir pattern in config")
		}
		pattern.DirPatternRegexp, err = regexp.Compile(pattern.DirPattern)
		if err != nil {
			return 0, err
		}
		if pattern.EpisodePattern == "" {
			return 0, fmt.Errorf("no episode pattern in config")
		}
		pattern.EpisodePatternRegexp, err = regexp.Compile(pattern.EpisodePattern)
		if err != nil {
			return 0, err
		}
		pattern.EpisodeFileAtLeastBytes, err = utils.SizeStringToBytesNum(pattern.EpisodeFileAtLeast)
		if err != nil {
			return 0, err
		}
		if pattern.SubtitlePattern != "" {
			pattern.SubtitlePatternRegexp, err = regexp.Compile(pattern.SubtitlePattern)
			if err != nil {
				return 0, err
			}
		}
	}
	p.patterns = cfg.Patterns
	return 0, nil
}

func (p *TvDir) IsDefaultEnable() bool {
	return true
}

type tvInfo struct {
	name          string
	year          *int
	tmdbid        *int
	season        *int
	originalName  string
	mediaFiles    map[episodeKey]*dirinfo.File
	subtitleFiles map[subtitleKey]*dirinfo.File
}

func (p *TvDir) Parse(entry *dirinfo.Entry, opts *parser.ParserMgrRunOpts) (ok bool, err error) {
	if entry.Type != dirinfo.DirEntry {
		return false, nil
	}
	if len(entry.FileList) <= 0 {
		return false, fmt.Errorf("no files in dir, entry: %s", entry.Name())
	}
	tvTargetDir, ok := opts.MediaTypeDirs[common.MediaTypeTv]
	if !ok {
		return false, fmt.Errorf("no tv target dir")
	}
	trashDir, ok := opts.MediaTypeDirs[common.MediaTypeTrash]
	if !ok {
		return false, fmt.Errorf("no trash dir")
	}
	info, err := p.parse(entry)
	if err != nil {
		return false, fmt.Errorf("parse error: %v", err)
	}
	if info == nil {
		return false, nil
	}
	level.Info(p.logger).Log("msg", "parsed", "dir", entry.Name(), "originalName", info.originalName, "year", info.year, "tmdbid", info.tmdbid)
	diskService := parser.GetDefaultDiskService()
	for mKey, file := range info.mediaFiles {
		err = diskService.RenameTvEpisode(&disk.TvEpisodeRenameTask{
			OldPath:      filepath.Join(entry.MotherPath, file.RelPathToMother),
			NewMotherDir: tvTargetDir,
			OriginalName: info.originalName,
			Year:         *info.year,
			Tmdbid:       *info.tmdbid,
			Season:       mKey.season,
			Episode:      mKey.episode,
		})
		if err != nil {
			return false, fmt.Errorf("rename tv episode error: %v", err)
		}
	}
	for sKey, file := range info.subtitleFiles {
		err = diskService.RenameTvSubtitle(&disk.TvSubtitleRenameTask{
			OldPath:      filepath.Join(entry.MotherPath, file.RelPathToMother),
			NewMotherDir: tvTargetDir,
			OriginalName: info.originalName,
			Year:         *info.year,
			Tmdbid:       *info.tmdbid,
			Season:       sKey.season,
			Episode:      sKey.episode,
			Language:     sKey.tag,
		})
		if err != nil {
			return false, fmt.Errorf("rename tv subtitle error: %v", err)
		}
	}
	err = diskService.MoveToTrash(&disk.MoveToTrashTask{
		Path:     filepath.Join(entry.MotherPath, entry.Name()),
		TrashDir: trashDir,
	})
	if err != nil {
		level.Warn(p.logger).Log("msg", "move to trash error", "err", err)
	}
	return true, nil
}

func (p *TvDir) parse(entry *dirinfo.Entry) (info *tvInfo, err error) {
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

type episodeKey struct {
	season  int
	episode int
}

type subtitleKey struct {
	tag     string
	season  int
	episode int
}

func (p *TvDir) matchPattern(entry *dirinfo.Entry, pattern *Pattern) (info *tvInfo, err error) {
	groups := pattern.DirPatternRegexp.FindStringSubmatch(entry.Name())
	if len(groups) <= 0 {
		return nil, nil
	}
	info = &tvInfo{
		tmdbid: pattern.Tmdbid,
		season: pattern.Season,
	}
	for i, name := range pattern.DirPatternRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "name":
			info.name = groups[i]
		case "year":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.year = &n
		case "tmdbid":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.tmdbid = &n
		case "season":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.season = &n
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	mediaFiles := make(map[episodeKey]*dirinfo.File)
	mediaFilesRev := make(map[string]episodeKey) // media file name without ext -> key
	subtitleFiles := make(map[subtitleKey]*dirinfo.File)
	// find all media files with no dup key
	for _, file := range entry.FileList {
		if !utils.IsMediaExt(file.Ext) || !utils.FileAtLeast(file, pattern.EpisodeFileAtLeastBytes) {
			continue
		}
		mKey, err := p.matchMediaFile(file, pattern, info)
		if err != nil {
			return nil, err
		}
		if mKey == nil {
			continue
		}
		if mKey.season < 0 || mKey.episode < 0 {
			continue
		}
		if _, ok := mediaFiles[*mKey]; ok {
			return nil, fmt.Errorf("duplicate media file key: %v", mKey)
		}
		mediaFiles[*mKey] = file
		fileNameWithoutExt := file.Name[:len(file.Name)-len(file.Ext)-1]
		mediaFilesRev[fileNameWithoutExt] = *mKey
	}
	if pattern.SubtitlePatternRegexp != nil { // subtitle is optional
		for _, file := range entry.FileList {
			if !utils.IsSubtitleExt(file.Ext) {
				continue
			}
			fileNameWithoutExt := file.Name[:len(file.Name)-len(file.Ext)-1]
			mKey, ok := mediaFilesRev[fileNameWithoutExt]
			if ok { // same name subtitle file see as media file's default subtitle
				sKey := &subtitleKey{tag: "", season: mKey.season, episode: mKey.episode}
				subtitleFiles[*sKey] = file
				continue
			}
			sKey, err := p.matchSubtitleFile(file, pattern, info)
			if err != nil {
				return nil, err
			}
			if sKey == nil {
				continue
			}
			if sKey.season < 0 || sKey.episode < 0 {
				continue
			}
			subtitleFiles[*sKey] = file
		}
	}
	if len(mediaFiles) <= 0 && len(subtitleFiles) <= 0 { // nothing to do, return no err but also no result
		return nil, nil
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid == nil {
		searchOpts := common.DefaultTmdbSearchOpts
		if info.year != nil {
			searchOpts["year"] = strconv.Itoa(*info.year)
		}
		results, err := tmdbService.GetSearchTVShow(info.name, searchOpts)
		if err != nil {
			return nil, err
		}
		if results.TotalResults <= 0 {
			return nil, fmt.Errorf("no tv show found, name: %s, year: %d", info.name, info.year)
		}
		if results.TotalResults > 1 {
			var hits []string
			for i := 0; i < 3 && i < len(results.Results); i++ {
				hits = append(hits, fmt.Sprintf("%s-%d", results.Results[i].OriginalName, results.Results[i].ID))
			}
			return nil, fmt.Errorf("multiple tvs found: %v", hits)
		}
		tmdbid := int(results.Results[0].ID)
		info.tmdbid = &tmdbid
	}
	detail, err := tmdbService.GetTVDetails(*info.tmdbid, common.DefaultTmdbSearchOpts)
	if err != nil {
		return nil, err
	}
	info.originalName = detail.OriginalName
	dt, err := common.ParseTmdbDateStr(detail.FirstAirDate)
	if err != nil {
		return nil, err
	}
	info.year = &dt.Year
	info.mediaFiles = mediaFiles
	info.subtitleFiles = subtitleFiles
	return info, nil
}

func (p *TvDir) matchMediaFile(file *dirinfo.File, pattern *Pattern, ti *tvInfo) (key *episodeKey, err error) {
	groups := pattern.EpisodePatternRegexp.FindStringSubmatch(file.PureName)
	if len(groups) <= 0 {
		return nil, nil
	}
	season := ti.season
	var episode *int
	for i, name := range pattern.EpisodePatternRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "season":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			season = &n
		case "episode":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			episode = &n
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	if season == nil || episode == nil {
		return nil, fmt.Errorf("no season or episode found in media file: %s", file.Name)
	}
	key = &episodeKey{season: *season, episode: *episode}
	return key, nil
}

func (p *TvDir) matchSubtitleFile(file *dirinfo.File, pattern *Pattern, ti *tvInfo) (key *subtitleKey, err error) {
	groups := pattern.SubtitlePatternRegexp.FindStringSubmatch(file.Name)
	if len(groups) <= 0 {
		return nil, nil
	}
	season := ti.season
	var episode *int
	tag := ""
	for i, name := range pattern.DirPatternRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "tag":
			tag = groups[i]
		case "season":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			season = &n
		case "episode":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			episode = &n
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	if season == nil || episode == nil {
		return nil, fmt.Errorf("no season or episode found in subtitle file: %s", file.Name)
	}
	key = &subtitleKey{tag: tag, season: *season, episode: *episode}
	return key, nil
}
