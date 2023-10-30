package ytsmovie

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/fileinfo"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
	"strconv"
)

type YTSMovieFileMatcher struct {
	tmdbClient       tmdbhttp.TMDBClient
	originNameRegexp *regexp.Regexp
	tmdbIDRegepx     *regexp.Regexp
	renamer          renamer.Renamer
	targetPath       string
}

var _ (matcher.Matcher) = (*YTSMovieFileMatcher)(nil)

const (
	originNameRegexGroupNum = 3
	originNameGroupName     = "name"
	originNameGroupYear     = "year"
	originNameRegexStr      = `^(?P<name>.*) \((?P<year>\d{4})\) .*YTS.*`

	tmdbIDNameRegexGroupNum = 2
	tmdbIDGroupName         = "tmdbid"
	tmdbIDRegexStr          = `.* movie tmdbid-(?P<tmdbid>\d*)$`
)

func NewYTSMovieDirMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*YTSMovieFileMatcher, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("tmdbClient nil")
	}
	if renamer == nil {
		return nil, fmt.Errorf("renamer nil")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("targetPath empty")
	}
	ytsMth := &YTSMovieFileMatcher{
		originNameRegexp: regexp.MustCompile(originNameRegexStr),
		tmdbIDRegepx:     regexp.MustCompile(tmdbIDRegexStr),
		tmdbClient:       tmdbClient,
		renamer:          renamer,
		targetPath:       targetPath,
	}
	return ytsMth, nil
}

func (mather *YTSMovieFileMatcher) Match(
	info *common.Info,
) (bool, error) {
	return mather.match(info)
}

func (mather *YTSMovieFileMatcher) match(
	info *common.Info,
) (bool, error) {
	var err error
	if info == nil {
		return false, fmt.Errorf("info nil")
	}
	if mather.tmdbClient == nil {
		return false, fmt.Errorf("tmdbClient nil")
	}
	dirInfo := info.Subs[0]
	if !dirInfo.IsDir {
		return false, nil
	}

	// match dir name
	groups := mather.originNameRegexp.FindStringSubmatch(dirInfo.Name)
	if len(groups) != originNameRegexGroupNum {
		return false, nil
	}
	var searthName string
	var year int
	var optionalTMDBID int64
	for i, name := range mather.originNameRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case originNameGroupName:
			searthName = groups[i]
		case originNameGroupYear:
			year, err = strconv.Atoi(groups[i])
			if err != nil {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unexpected name:%s", name)
		}
	}

	// try to get tmdbid if has
	if groups = mather.tmdbIDRegepx.FindStringSubmatch(dirInfo.Name); len(groups) == tmdbIDNameRegexGroupNum {
		for i, name := range mather.tmdbIDRegepx.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case tmdbIDGroupName:
				optionalTMDBID, err = strconv.ParseInt(groups[i], 10, 64)
				if err != nil {
					optionalTMDBID = 0
				}
			default:
				return false, fmt.Errorf("unexpected name:%s", name)
			}
		}
	}

	// search tmdb by id or by name
	var matched *common.MatchedMovie
	if optionalTMDBID != 0 {
		data, err := tmdbhttp.SearchMovieByTmdbID(mather.tmdbClient, optionalTMDBID)
		if err != nil {
			return false, TMDBErr{err: err}
		}
		matched, err = tmdbhttp.ConvMovie(data)
		if err != nil {
			return false, TMDBErr{err: err}
		}
	} else {
		data, err := tmdbhttp.SearchOnlyOneMovieByName(mather.tmdbClient, searthName, year)
		if err != nil {
			return false, TMDBErr{err: err}
		}
		matched, err = tmdbhttp.ConvMovie(data)
		if err != nil {
			return false, TMDBErr{err: err}
		}
	}
	if matched == nil {
		return false, nil
	}

	// rename's records
	var renames []renamer.RenameRecord

	// if has one and only one movie file, rename it
	if mediaFiles := fileinfo.GetBiggerThanMediaFiles(movieFileMinSize, info); len(mediaFiles) == 1 {
		movieFile := mediaFiles[0]
		movieOldPath := renamer.Path{info.DirPath}
		movieOldPath = append(movieOldPath, movieFile.Paths...)
		movieOldPath = append(movieOldPath, fmt.Sprintf("%s%s", movieFile.Name, movieFile.Ext))
		movieNewPath, err := renamer.TargetMovieFilePath(matched, mather.targetPath, movieFile.Ext)
		if err != nil {
			return false, err
		}
		renames = append(renames, renamer.RenameRecord{Old: movieOldPath, New: movieNewPath})
		movieFile.State = 1

		// if has one and only one same name subtitle file
		if subtitles := fileinfo.GetSameNameSubtitleFiles(movieFile.Name, info); len(subtitles) == 1 {
			subtitleFile := subtitles[0]
			subtileOldPath := renamer.Path{info.DirPath}
			subtileOldPath = append(subtileOldPath, subtitleFile.Paths...)
			subtileOldPath = append(subtileOldPath, fmt.Sprintf("%s%s", subtitleFile.Name, subtitleFile.Ext))
			subtileNewPath, err := renamer.TargetMovieFilePath(matched, mather.targetPath, subtitleFile.Ext)
			if err != nil {
				return false, err
			}
			renames = append(renames, renamer.RenameRecord{Old: subtileOldPath, New: subtileNewPath})
			subtitleFile.State = 1
		}
	} else if len(mediaFiles) > 1 {
		return false, nil
	}

	// if has multiple subtile files, stop process, NOT support now
	if subtiles := fileinfo.GetAllSubtitleFiles(info); len(subtiles) > 1 {
		for _, subtitleFile := range subtiles {
			if mappingTo, ok := subtitleMapping[subtitleFile.Name]; ok {
				subtileOldPath := renamer.Path{info.DirPath}
				subtileOldPath = append(subtileOldPath, subtitleFile.Paths...)
				subtileOldPath = append(subtileOldPath, fmt.Sprintf("%s%s", subtitleFile.Name, subtitleFile.Ext))
				subtileNewPath, err := renamer.TargetMovieSubtitleFilePath(matched, mather.targetPath, mappingTo, subtitleFile.Ext)
				if err != nil {
					return false, err
				}
				renames = append(renames, renamer.RenameRecord{Old: subtileOldPath, New: subtileNewPath})
				subtitleFile.State = 1
			}
		}
	}

	// rename
	err = mather.renamer.Rename(renames)
	if err != nil {
		return false, err
	}

	// check and post delete useless dir
	_ = common.PostClean(info)

	return true, nil
}

var subtitleMapping = map[string]string{
	"English":                   "en",
	"Simplified.chi":            "zh_CN",
	"Traditional.chi":           "zh_HK",
	"Chinese Simplified.chi":    "zh_CN",
	"Chinese Traditional.chi":   "zh_HK",
	"Japanese.jpn":              "ja_JP",
	"Korean.kor":                "ko_KR",
	"Simplified Chinese.chi":    "zh_CN",
	"Traditional Chinese.chi":   "zh_HK",
	"Chinese (Simplified).chi":  "zh_CN",
	"Chinese (Traditional).chi": "zh_HK",
}

const (
	movieFileMinSize = 1 * 1024 * 1024
)

type TMDBErr struct {
	err error
}

func (e TMDBErr) Error() string {
	return fmt.Sprintf("tmdb search err:%s", e.err)
}

func (e TMDBErr) UnWrap() error {
	return e.err
}

type RenameErr struct {
	err error
}

func (e RenameErr) Error() string {
	return fmt.Sprintf("rename err:%s", e.err)
}

func (e RenameErr) UnWrap() error {
	return e.err
}
