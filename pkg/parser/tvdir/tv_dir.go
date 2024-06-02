package tvdir

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
	name = "tvdir"
)

func init() {
	parser.RegisterParser(name, &TvDir{})
}

type TvDir struct {
	logger   log.Logger
	patterns []*Pattern
}

type Pattern struct {
	DirPatternStr      string `toml:"dir_pattern"`
	EpisodePatternStr  string `toml:"episode_pattern"`
	EpisodeFileAtLeast string `toml:"episode_file_at_least"`
	SubtitlePatternStr string `toml:"subtitle_pattern"`
	Season             *int   `toml:"season"`

	DirPattern              *regexp.Regexp
	EpisodePattern          *regexp.Regexp
	EpisodeFileAtLeastBytes int64
	SubtitlePattern         *regexp.Regexp
}

type Config struct {
	Patterns []*Pattern `toml:"patterns"`
}

func (p *TvDir) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
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
		pattern.EpisodePattern, err = regexp.Compile(pattern.EpisodePatternStr)
		if err != nil {
			return 0, err
		}
		pattern.EpisodeFileAtLeastBytes, err = utils.SizeStringToBytesNum(pattern.EpisodeFileAtLeast)
		if err != nil {
			return 0, err
		}
		pattern.SubtitlePattern, err = regexp.Compile(pattern.SubtitlePatternStr)
		if err != nil {
			return 0, err
		}
	}
	p.patterns = cfg.Patterns
	return 0, nil
}

func (p *TvDir) IsDefaultEnable() bool {
	return true
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
			Year:         info.year,
			Tmdbid:       info.tmdbid,
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
			Year:         info.year,
			Tmdbid:       info.tmdbid,
			Season:       sKey.season,
			Episode:      sKey.episode,
			Language:     sKey.lang,
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

type tvInfo struct {
	name          string
	year          int
	tmdbid        int
	season        int
	originalName  string
	mediaFiles    map[episodeKey]*dirinfo.File
	subtitleFiles map[subtitleKey]*dirinfo.File
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
	lang    string
	season  int
	episode int
}

func (p *TvDir) matchPattern(entry *dirinfo.Entry, pattern *Pattern) (info *tvInfo, err error) {
	groups := pattern.DirPattern.FindStringSubmatch(entry.Name())
	if len(groups) <= 0 {
		return nil, nil
	}
	info = &tvInfo{}
	for i, name := range pattern.DirPattern.SubexpNames() {
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
			info.year = n
		case "tmdbid":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.tmdbid = n
		case "season":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			info.season = n
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	mediaFiles := make(map[episodeKey]*dirinfo.File)
	mediaFileRev := make(map[string]*episodeKey)
	subtitleFiles := make(map[subtitleKey]*dirinfo.File)
	for _, file := range entry.FileList {
		if utils.IsMediaExt(file.Ext) && utils.FileAtLeast(file, pattern.EpisodeFileAtLeastBytes) {
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
			mediaFiles[*mKey] = file
			fileNameWithoutExt := file.Name[:len(file.Name)-len(file.Ext)-1]
			mediaFileRev[fileNameWithoutExt] = mKey
		}
	}
	for _, file := range entry.FileList {
		if utils.IsSubtitleExt(file.Ext) {
			fileNameWithoutExt := file.Name[:len(file.Name)-len(file.Ext)-1]
			mKey, ok := mediaFileRev[fileNameWithoutExt]
			if ok {
				sKey := &subtitleKey{lang: "", season: mKey.season, episode: mKey.episode}
				subtitleFiles[*sKey] = file
				continue
			}
			if pattern.SubtitlePatternStr == "" {
				continue
			}
			sKey, err := p.matchSubtitleFile(file, pattern)
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
	if len(mediaFiles) <= 0 && len(subtitleFiles) <= 0 {
		return nil, nil
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid <= 0 {
		searchOpts := common.DefaultTmdbSearchOpts
		if info.year >= common.ValidStartYear {
			searchOpts["year"] = strconv.Itoa(info.year)
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
		info.tmdbid = int(results.Results[0].ID)
	}
	detail, err := tmdbService.GetTVDetails(info.tmdbid, common.DefaultTmdbSearchOpts)
	if err != nil {
		return nil, err
	}
	info.originalName = detail.OriginalName
	dt, err := common.ParseTmdbDateStr(detail.FirstAirDate)
	if err != nil {
		return nil, err
	}
	info.year = dt.Year
	info.mediaFiles = mediaFiles
	info.subtitleFiles = subtitleFiles
	return info, nil
}

func (p *TvDir) matchMediaFile(file *dirinfo.File, pattern *Pattern, tvInfo *tvInfo) (key *episodeKey, err error) {
	groups := pattern.EpisodePattern.FindStringSubmatch(file.PureName)
	if len(groups) <= 0 {
		return nil, nil
	}
	key = &episodeKey{season: -1, episode: -1}
	if pattern.Season != nil && *pattern.Season >= 0 {
		key.season = *pattern.Season
	}
	if tvInfo.season >= 0 {
		key.season = tvInfo.season
	}
	for i, name := range pattern.EpisodePattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "season":
			season, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			key.season = season
		case "episode":
			episode, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			key.episode = episode
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	return key, nil
}

func (p *TvDir) matchSubtitleFile(file *dirinfo.File, pattern *Pattern) (key *subtitleKey, err error) {
	groups := pattern.SubtitlePattern.FindStringSubmatch(file.Name)
	if len(groups) <= 0 {
		return nil, nil
	}
	key = &subtitleKey{lang: "", season: -1, episode: -1}
	for i, name := range pattern.DirPattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "lang":
			key.lang = groups[i]
		case "season":
			season, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			key.season = season
		case "episode":
			episode, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, err
			}
			key.episode = episode
		default:
			return nil, fmt.Errorf("unknown group name: %s", name)
		}
	}
	return key, nil
}
