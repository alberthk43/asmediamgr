package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Info struct {
	DirPath string
	Subs    []Single
}

type Single struct {
	Paths []string
	Name  string
	Ext   string
	IsDir bool
	Size  int64
	State int
}

type MatchedCommon struct {
	OriginalTitle    string
	OriginalLanguage string
	TmdbID           int64
	Adult            bool
	Year             int32
}

type MatchedMovie struct {
	MatchedCommon
}

type MatchedTV struct {
	MatchedCommon
	Season int32
	EpNum  int32
}

var mediaFileExts = map[string]interface{}{
	".mp4":  struct{}{},
	".mkv":  struct{}{},
	".ts":   struct{}{},
	".rmvb": struct{}{},
}

func IsMediaFile(ext string) bool {
	_, ok := mediaFileExts[ext]
	return ok
}

var subtitleFileExts = map[string]interface{}{
	".srt": struct{}{},
	".ass": struct{}{},
}

func IsSubtitleFile(ext string) bool {
	_, ok := subtitleFileExts[ext]
	return ok
}

func PostClean(info *Info) error {
	if info == nil {
		return fmt.Errorf("info nil")
	}
	isAllDeleted := true
	for _, one := range info.Subs {
		if one.IsDir {
			continue
		}
		if one.State > 0 {
			continue
		}
		if _, ok := uselessExt[one.Ext]; !ok {
			isAllDeleted = false
			continue
		}
		fullPath := filepath.Join(info.DirPath)
		for _, path := range one.Paths {
			fullPath = filepath.Join(fullPath, path)
		}
		log.Printf("todelete %+v\n", one)
		fileName := fmt.Sprintf("%s%s", one.Name, one.Ext)
		fullPath = filepath.Join(fullPath, fileName)
		err := os.Remove(fullPath)
		if err != nil {
			return err
		}
	}
	if isAllDeleted {
		dir := info.Subs[0]
		fullPath := filepath.Join(info.DirPath, dir.Name)
		log.Printf("todelete %+v\n", dir)
		err := os.Remove(fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

var uselessExt = map[string]interface{}{
	".nfo":     nil,
	".torrent": nil,
	".txt":     nil,
	".jpg":     nil,
}
