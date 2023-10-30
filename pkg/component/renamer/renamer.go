package renamer

import (
	"asmediamgr/pkg/common"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Path []string

type RenameRecord struct {
	Old, New Path
}

type Renamer interface {
	Rename(records []RenameRecord) error
}

type FileRenamer struct{}

var _ Renamer = (*FileRenamer)(nil)

func (renamer *FileRenamer) Rename(records []RenameRecord) error {
	for _, v := range records {
		err := renamer.renameOne(v.Old, v.New)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*FileRenamer) renameOne(old, new Path) error {
	log.Printf("renamer old:\n")
	for _, v := range old {
		log.Printf("**** %s\n", v)
	}
	log.Printf("renamer new:\n")
	for _, v := range new {
		log.Printf("**** %s\n", v)
	}

	if len(new) == 0 {
		return fmt.Errorf("new empty")
	}
	targetDir := new[0]
	stat, err := os.Stat(targetDir)
	if err != nil {
		return err
	}
	statMode := stat.Mode()
	targetPath := filepath.Join(new[:len(new)-1]...)
	err = os.MkdirAll(targetPath, statMode)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(old...), filepath.Join(new...))
	if err != nil {
		return err
	}
	return nil
}

func TargetMovieFilePath(
	matched *common.MatchedMovie,
	targetDir string,
	ext string,
) (Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	path = append(path, fmt.Sprintf("%s (%.4d)%s", escapedTitle, matched.Year, ext))
	return path, nil
}

func TargetMovieShortFilePath(
	matched *common.MatchedMovie,
	targetDir string,
	publisherName string,
	ext string,
) (Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := Path{targetDir}
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
) (Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("movie nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := Path{targetDir}
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

func ReplaceSpecialCharacters(str string) string {
	for k, v := range specialCharacterMapping {
		str = strings.ReplaceAll(str, k, v)
	}
	return str
}

func TargetTVEpFilePath(
	matched *common.MatchedTV,
	targetDir string,
	ext string,
) (Path, error) {
	if matched == nil {
		return nil, fmt.Errorf("tv nil")
	}
	if ext == "" {
		return nil, fmt.Errorf("ext empty")
	}
	escapedTitle := ReplaceSpecialCharacters(matched.OriginalTitle)
	path := Path{targetDir}
	path = append(path, fmt.Sprintf("%s (%.4d) [tmdbid-%d]", escapedTitle, matched.Year, matched.TmdbID))
	path = append(path, fmt.Sprintf("Season %d", matched.Season))
	path = append(path, fmt.Sprintf("S%0.2dE%0.2d%s", matched.Season, matched.EpNum, ext))
	return path, nil
}
