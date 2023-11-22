package stats

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type MediaTp int

const (
	Unknown MediaTp = iota
	Movie
)

type Media struct {
	MediaType    MediaTp
	TmdbID       int64
	OriginalName string
	Year         int32
}

type MediaStats struct {
	Medias map[MediaTp][]Media
}

func newMediaStats() *MediaStats {
	ms := &MediaStats{
		Medias: make(map[MediaTp][]Media),
	}
	return ms
}

func GatherMovieStats(moviePath string) (*MediaStats, error) {
	ms := newMediaStats()
	dir, err := os.Open(moviePath)
	if err != nil {
		return nil, err
	}
	subs, err := dir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	for _, sub := range subs {
		if !sub.IsDir() {
			continue
		}
		dirName := sub.Name()
		m, err := parseMovieInfoFromName(dirName)
		if err != nil {
			continue
		}
		movies := ms.Medias[m.MediaType]
		movies = append(movies, *m)
		ms.Medias[m.MediaType] = movies
	}
	return ms, nil
}

var (
	regexMovieName = regexp.MustCompile(`^(?P<name>.*) \((?P<year>\d+)\) \[tmdbid-(?P<tmdbid>\d+)\].*$`)
)

func parseMovieInfoFromName(dirName string) (*Media, error) {
	groups := regexMovieName.FindStringSubmatch(dirName)
	names := regexMovieName.SubexpNames()
	if len(groups) != len(names) {
		return nil, fmt.Errorf("invalid movie dir name:%s", dirName)
	}
	var originalName string
	var year int32
	var tmdbid int64
	for i := 0; i < len(names); i++ {
		if i == 0 {
			continue
		}
		group := groups[i]
		name := names[i]
		switch name {
		case "name":
			originalName = group
		case "year":
			n, err := strconv.ParseInt(group, 10, 31)
			if err != nil {
				return nil, fmt.Errorf("year parse err:%s", err)
			}
			if n < 1000 || n > 3000 {
				return nil, fmt.Errorf("invalid year:%d", n)
			}
			year = int32(n)
		case "tmdbid":
			n, err := strconv.ParseInt(group, 10, 63)
			if err != nil {
				return nil, fmt.Errorf("tmdbid parse err:%s", err)
			}
			tmdbid = n
		}
	}
	m := &Media{
		MediaType:    Movie,
		TmdbID:       tmdbid,
		OriginalName: originalName,
		Year:         year,
	}
	return m, nil
}
