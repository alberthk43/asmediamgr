package javfc

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/fileinfo"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
	"strconv"
)

const (
	fcIDGroupName = "fcid"

	fileNameRegexStr = `.*(FC2-PPV|FC2PPV)-(?P<fcid>\d+).*`

	minMediaSize = 300 * 1024 * 1024
)

type JavFCMatcher struct {
	fileNameRegex *regexp.Regexp
	renamer       renamer.Renamer
	targetPath    string
}

func NewJavFCMatcher(
	renamer renamer.Renamer,
	targetPath string,
) (*JavFCMatcher, error) {
	matcher := &JavFCMatcher{
		fileNameRegex: regexp.MustCompile(fileNameRegexStr),
		renamer:       renamer,
		targetPath:    targetPath,
	}
	return matcher, nil
}

var _ matcher.Matcher = (*JavFCMatcher)(nil)

func (matcher *JavFCMatcher) Match(info *common.Info) (bool, error) {
	if groups := matcher.fileNameRegex.FindStringSubmatch(info.Subs[0].Name); len(groups) == 0 {
		return false, fmt.Errorf("not a fc dir")
	}
	fcID := int64(0)
	var mediaFile *common.Single
	if mediaFiles := fileinfo.GetBiggerThanMediaFiles(minMediaSize, info); len(mediaFiles) == 1 {
		curFile := mediaFiles[0]
		if groups := matcher.fileNameRegex.FindStringSubmatch(curFile.Name); len(groups) > 0 {
			for i, name := range matcher.fileNameRegex.SubexpNames() {
				switch name {
				case fcIDGroupName:
					n, err := strconv.ParseInt(groups[i], 10, 64)
					if err != nil {
						return false, err
					}
					fcID = n
					mediaFile = curFile
				}
			}
		}
	} else if len(mediaFiles) > 1 {
		return false, fmt.Errorf("too many media files, not support not")
	}
	if fcID <= 0 {
		return false, fmt.Errorf("not found fc id")
	}
	if mediaFile == nil {
		return false, fmt.Errorf("not found media file")
	}
	oldPath := renamer.Path{info.DirPath}
	oldPath = append(oldPath, mediaFile.Paths...)
	oldPath = append(oldPath, fmt.Sprintf("%s%s", mediaFile.Name, mediaFile.Ext))
	newPath, err := buildFcFileNewPath(mediaFile, fcID, matcher.targetPath, mediaFile.Ext)
	if err != nil {
		return false, err
	}
	renames := []renamer.RenameRecord{{Old: oldPath, New: newPath}}
	err = matcher.renamer.Rename(renames)
	if err != nil {
		return false, err
	}
	_ = common.PostClean(info)
	return true, nil
}

func buildFcFileNewPath(
	mediaFile *common.Single,
	fcID int64,
	targetDir string,
	ext string) (renamer.Path, error) {
	path := renamer.Path{targetDir}
	path = append(path, fmt.Sprintf("FC2-PPV-%d", fcID))
	path = append(path, fmt.Sprintf("FC2-PPV-%d%s", fcID, ext))
	return path, nil
}
