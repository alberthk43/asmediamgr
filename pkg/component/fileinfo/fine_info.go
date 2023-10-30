package fileinfo

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/tmdbhttp"
	"fmt"
	"regexp"
	"strconv"
)

func GetBiggerThanMediaFiles(
	size int64,
	info *common.Info,
) (res []*common.Single) {
	if info == nil {
		return
	}
	for i := 0; i < len(info.Subs); i++ {
		sub := &info.Subs[i]
		if sub.IsDir {
			continue
		}
		if sub.Size < size {
			continue
		}
		if !common.IsMediaFile(sub.Ext) {
			continue
		}
		res = append(res, sub)
	}
	return
}

func GetSameNameSubtitleFiles(
	name string,
	info *common.Info,
) (res []*common.Single) {
	if info == nil {
		return
	}
	if name == "" {
		return
	}
	for i := 0; i < len(info.Subs); i++ {
		sub := &info.Subs[i]
		if sub.IsDir {
			continue
		}
		if name != sub.Name {
			continue
		}
		if !common.IsSubtitleFile(sub.Ext) {
			continue
		}
		res = append(res, sub)
	}
	return
}

func GetAllSubtitleFiles(
	info *common.Info,
) (res []*common.Single) {
	if info == nil {
		return
	}
	for i := 0; i < len(info.Subs); i++ {
		sub := &info.Subs[i]
		if sub.IsDir {
			continue
		}
		if !common.IsSubtitleFile(sub.Ext) {
			continue
		}
		res = append(res, sub)
	}
	return
}

// CheckIsSingleMediaFile return err if info is NOT a single media type file
func CheckIsSingleMediaFile(
	info *common.Info,
) error {
	if err := CheckIsSingleFile(info); err != nil {
		return err
	}
	if !common.IsMediaFile(info.Subs[0].Ext) {
		return fmt.Errorf("not media file")
	}
	return nil
}

// CheckIsSingleFile return err if info is NOT a single file
func CheckIsSingleFile(
	info *common.Info,
) error {
	if info == nil {
		return fmt.Errorf("info nil")
	}
	if len(info.Subs) != 1 {
		return fmt.Errorf("not signle file")
	}
	fileInfo := &info.Subs[0]
	if fileInfo.IsDir {
		return fmt.Errorf("not signle file")
	}
	return nil
}

// CheckIsDir return err if info is NOT a dir list
func CheckIsDir(
	info *common.Info,
) error {
	if info == nil {
		return fmt.Errorf("info nil")
	}
	if len(info.Subs) <= 1 {
		return fmt.Errorf("not dir")
	}
	dirInfo := &info.Subs[0]
	if !dirInfo.IsDir {
		return fmt.Errorf("not dir")
	}
	return nil
}

func MatchName(
	raw string,
	regex *regexp.Regexp,
) (name string, err error) {
	if raw == "" {
		return "", fmt.Errorf("name emtpty")
	}
	if regex == nil {
		return "", fmt.Errorf("regex is nil")
	}
	groups := regex.FindStringSubmatch(raw)
	for i, groupName := range regex.SubexpNames() {
		if i == 0 {
			continue
		}
		switch groupName {
		case "name":
			name = groups[i]
		}
	}
	return name, nil
}

func MatchTMDBID(
	raw string,
	regex *regexp.Regexp,
) (tmdbID int64, err error) {
	if raw == "" {
		return 0, fmt.Errorf("raw empty")
	}
	if regex == nil {
		return 0, fmt.Errorf("regex is nil")
	}
	groups := regex.FindStringSubmatch(raw)
	for i, groupName := range regex.SubexpNames() {
		if i == 0 {
			continue
		}
		switch groupName {
		case "tmdbid":
			str := groups[i]
			n, err := strconv.ParseInt(str, 10, 63)
			if err != nil {
				return 0, fmt.Errorf("tmdbid not int")
			}
			tmdbID = n
		default:
			continue
		}
	}
	return tmdbID, nil
}

var (
	movieTMDBIDRegex = regexp.MustCompile(`.* movie tmdbid-(?P<tmdbid>\d+)$`)
	tvTMDBIDRegex    = regexp.MustCompile(`.* tv tmdbid-(?P<tmdbid>\d+)$`)
)

func MatchMovieTMDBID(
	raw string,
) (tmdbID int64, err error) {
	return MatchTMDBID(raw, movieTMDBIDRegex)
}

func MatchTvTMDBID(
	raw string,
) (tmdbID int64, err error) {
	return MatchTMDBID(raw, tvTMDBIDRegex)
}

func SearchTv(
	tmdbID int64,
	searchName string,
	tmdbClient tmdbhttp.TMDBClient,
) (matched *common.MatchedTV, err error) {
	if tmdbID > 0 {
		data, err := tmdbhttp.SearchTVByTmdbID(tmdbClient, tmdbID)
		if err != nil {
			return nil, err
		}
		matched, err = tmdbhttp.ConvTV(data)
		if err != nil {
			return nil, err
		}
	} else {
		if searchName == "" {
			return nil, fmt.Errorf("searchName not found")
		}
		data, err := tmdbhttp.SearchOnlyOneTVByName(tmdbClient, searchName)
		if err != nil {
			return nil, err
		}
		matched, err = tmdbhttp.ConvTV(data)
		if err != nil {
			return nil, err
		}
	}
	return matched, nil
}

func MatchTvEp(
	name string,
	defaultSeason int32,
	epRegex, optionalSeasonRegex *regexp.Regexp,
) (season, epNum int32, err error) {
	if name == "" {
		return 0, 0, fmt.Errorf("name empty")
	}
	if epRegex == nil {
		return 0, 0, fmt.Errorf("epRegex is nil")
	}
	season = defaultSeason
	if groups := epRegex.FindStringSubmatch(name); len(groups) > 0 {
		for i, groupName := range epRegex.SubexpNames() {
			if i == 0 {
				continue
			}
			switch groupName {
			case "season":
				str := groups[i]
				n, err := strconv.ParseInt(str, 10, 31)
				if err != nil {
					return 0, 0, fmt.Errorf("season not int")
				}
				season = int32(n)
			case "ep":
				str := groups[i]
				n, err := strconv.ParseInt(str, 10, 31)
				if err != nil {
					return 0, 0, fmt.Errorf("season not int")
				}
				epNum = int32(n)
			default:
				continue
			}
		}
	}
	if optionalSeasonRegex != nil {
		if groups := epRegex.FindStringSubmatch(name); len(groups) > 0 {
			for i, groupName := range epRegex.SubexpNames() {
				if i == 0 {
					continue
				}
				switch groupName {
				case "season":
					str := groups[i]
					n, err := strconv.ParseInt(str, 10, 31)
					if err != nil {
						return 0, 0, fmt.Errorf("season not int")
					}
					season = int32(n)
				default:
					continue
				}
			}
		}
	}
	return season, epNum, nil
}
