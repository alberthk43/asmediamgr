package renamehelper

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
	"fmt"
	"strings"
)

func BuildRenameRecordFromSubInfo(
	motherPathDir string,
	targetPathDir string,
	fileInfo *common.Single,
	tvInfo *common.MatchedTV,
	seasonNum int32,
	epNum int32,
) (*renamer.RenameRecord, error) {
	old := renamer.Path{motherPathDir}
	old = append(old, fileInfo.Paths...)
	oldFile := fmt.Sprintf("%s%s", fileInfo.Name, fileInfo.Ext)
	old = append(old, oldFile)
	new, err := TargetTVEpFilePath(tvInfo, targetPathDir, fileInfo.Ext)
	if err != nil {
		return nil, err
	}
	record := &renamer.RenameRecord{Old: old, New: new}
	return record, nil
}

func TargetMovieFilePath(
	matched *common.MatchedMovie,
	targetDir string,
	ext string,
) (renamer.Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := renamer.Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	path = append(path, fmt.Sprintf("%s (%.4d)%s", escapedTitle, matched.Year, ext))
	return path, nil
}

func TargetMovieShortFilePath(
	matched *common.MatchedMovie,
	targetDir string,
	publisherName string,
	ext string,
) (renamer.Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := renamer.Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	path = append(path, "shorts")
	if publisherName != "" {
		path = append(path, fmt.Sprintf("%s %s%s", escapedTitle, publisherName, ext))
	} else {
		path = append(path, fmt.Sprintf("short %s %s%s", escapedTitle, publisherName, ext))
	}
	return path, nil
}

func TargetMovieSubtitleFilePath(
	matched *common.MatchedMovie,
	targetDir string,
	lang string,
	ext string,
) (renamer.Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := renamer.Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	if lang != "" {
		path = append(path, fmt.Sprintf("%s (%.4d).%s%s", escapedTitle, matched.Year, lang, ext))
	} else {
		path = append(path, fmt.Sprintf("%s (%.4d)%s", escapedTitle, matched.Year, ext))
	}
	return path, nil
}

var specialCharacterMapping = map[string]string{
	`#`: ` `,
	// `%`: ` `,
	// `&`: ` `,
	// `{`: ` `,
	// `}`: ` `,
	`\`: ` `,
	// `$`: ` `,
	// `!`: ` `,
	`'`: ` `,
	// `"`: ` `,
	`:`: ` `,
	// `<`: ` `,
	// `>`: ` `,
	// `*`: ` `,
	// `?`: ` `,
	`/`: ` `,
	// `+`: ` `,
	"`": ` `,
	`|`: ` `,
	`=`: ` `,
}

func TargetTVEpFilePath(
	matched *common.MatchedTV,
	targetDir string,
	ext string,
) (renamer.Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("tv nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := renamer.Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	path = append(path, fmt.Sprintf("Season %d", matched.Season))
	path = append(path, fmt.Sprintf("S%0.2dE%0.2d%s", matched.Season, matched.EpNum, ext))
	return path, nil
}

func ReplaceSpecialCharacters(str string) string {
	for k, v := range specialCharacterMapping {
		str = strings.ReplaceAll(str, k, v)
	}
	return str
}

func TargetMovieDir(
	movieInfo *common.MatchedMovie,
) (dirName string) {
	dirName = fmt.Sprintf("%s (%.4d) [tmdbid-%d]", movieInfo.OriginalTitle, movieInfo.Year, movieInfo.TmdbID)
	return dirName
}

func TargetMovieShortDir() (shortDirName string) {
	return "shorts"
}
