package stat

import (
	"fmt"

	"github.com/albert43/asmediamgr/pkg/tmdb"
	"github.com/albert43/asmediamgr/pkg/utils"
)

type multipleMovieChecker struct{}

func (mmc *multipleMovieChecker) check(mStat *movieStat) StatErr {
	if len(mStat.movieFiles) > 1 {
		return &MultipleMovieStatErr{tmdbid: mStat.tmdbid, fileInfos: mStat.movieFiles}
	}
	return nil
}

type MultipleMovieStatErr struct {
	tmdbid    int
	fileInfos []*fileInfo
}

func (mmsErr *MultipleMovieStatErr) Error() string {
	return "multiple movie files"
}

func (mmsErr *MultipleMovieStatErr) toMarkdownContent() (string, error) {
	str := "## Multiple movie files:\n"
	if mmsErr.tmdbid > 0 {
		str += "\n"
		str += tmdb.BuildTmdbMovieLink(mmsErr.tmdbid)
		str += "\n\n"
	}
	for _, fileInfo := range mmsErr.fileInfos {
		str += fmt.Sprintf("  - %s\n", fileInfo.path)
	}
	str += "\n"
	return str, nil
}

type largetMovieChecker struct {
	sizeThreshold int64
}

func (lmc *largetMovieChecker) check(mStat *movieStat) StatErr {
	for _, file := range mStat.movieFiles {
		if file.size > lmc.sizeThreshold {
			return &LargeMovieStatErr{tmdbid: mStat.tmdbid, fileInfos: mStat.movieFiles}
		}
	}
	return nil
}

type LargeMovieStatErr struct {
	tmdbid    int
	fileInfos []*fileInfo
}

func (lmsErr *LargeMovieStatErr) Error() string {
	return "large movie files"
}

func (lmsErr *LargeMovieStatErr) toMarkdownContent() (string, error) {
	str := "## Large movie files:\n"
	if lmsErr.tmdbid > 0 {
		str += "\n"
		str += tmdb.BuildTmdbMovieLink(lmsErr.tmdbid)
		str += "\n\n"
	}
	for _, fileInfo := range lmsErr.fileInfos {
		str += fmt.Sprintf("  - %s size=%s\n", fileInfo.path, utils.BytesNumToSizeString(fileInfo.size))
	}
	str += "\n"
	return str, nil
}

type multipleTvEpisodeChecker struct{}

func (mtec *multipleTvEpisodeChecker) check(tvStat *tvStat) StatErr {
	statErr := &MultipleTvEpisodeStatErr{tmdbid: tvStat.tmdbid}
	for _, files := range tvStat.episodeFiles {
		if len(files) > 1 {
			statErr.fileInfos = append(statErr.fileInfos, files...)
		}
	}
	if len(statErr.fileInfos) > 0 {
		return statErr
	}
	return nil
}

type MultipleTvEpisodeStatErr struct {
	tmdbid    int
	fileInfos []*fileInfo
}

func (mtesErr *MultipleTvEpisodeStatErr) Error() string {
	return "multiple tv episode files"
}

func (mtesErr *MultipleTvEpisodeStatErr) toMarkdownContent() (string, error) {
	str := "## Multiple tv episode files:\n"
	if mtesErr.tmdbid > 0 {
		str += "\n"
		str += tmdb.BuildTmdbTvLink(mtesErr.tmdbid)
		str += "\n\n"
	}
	for _, fileInfo := range mtesErr.fileInfos {
		str += fmt.Sprintf("  - %s\n", fileInfo.path)
	}
	str += "\n"
	return str, nil
}

type largeTvEpisodeChecker struct {
	sizeThreshold int64
}

func (ltec *largeTvEpisodeChecker) check(tvStat *tvStat) StatErr {
	statErr := &LargeTvEpisodeStatErr{tmdbid: tvStat.tmdbid}
	for _, files := range tvStat.episodeFiles {
		for _, file := range files {
			if file.size > ltec.sizeThreshold {
				statErr.fileInfos = append(statErr.fileInfos, file)
			}
		}
	}
	if len(statErr.fileInfos) > 0 {
		return statErr
	}
	return nil
}

type LargeTvEpisodeStatErr struct {
	tmdbid    int
	fileInfos []*fileInfo
}

func (ltesErr *LargeTvEpisodeStatErr) Error() string {
	return "large tv episode files"
}

func (ltesErr *LargeTvEpisodeStatErr) toMarkdownContent() (string, error) {
	str := "## Large tv episode files:\n"
	if ltesErr.tmdbid > 0 {
		str += "\n"
		str += tmdb.BuildTmdbTvLink(ltesErr.tmdbid)
		str += "\n\n"
	}
	for _, fileInfo := range ltesErr.fileInfos {
		str += fmt.Sprintf("  - %s size=%s\n", fileInfo.path, utils.BytesNumToSizeString(fileInfo.size))
	}
	str += "\n"
	return str, nil
}
