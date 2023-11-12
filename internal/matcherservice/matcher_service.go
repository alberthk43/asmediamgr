package matcherservice

import (
	"asmediamgr/pkg/component/dirpath"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher"
	"asmediamgr/pkg/matcher/aniepfile"
	"asmediamgr/pkg/matcher/ddhdtv"
	"asmediamgr/pkg/matcher/huanyue"
	"asmediamgr/pkg/matcher/javfc"
	"asmediamgr/pkg/matcher/megusta"
	"asmediamgr/pkg/matcher/nekomoe"
	"asmediamgr/pkg/matcher/preregex"
	"asmediamgr/pkg/matcher/sevenacg"
	"asmediamgr/pkg/matcher/shorts"
	"asmediamgr/pkg/matcher/singlemoviefile"
	"asmediamgr/pkg/matcher/tvepfile"
	"asmediamgr/pkg/matcher/ytsmovie"
	"os"
)

var matcherMgr *matcher.MatcherMgr

func GetMatcherMgr() *matcher.MatcherMgr {
	return matcherMgr
}

func init() {
	matcherMgr = matcher.NewMatchMgr()
}

func InitMatcher(moviePath, tvPath, javPath string) error {
	apiKey := os.Getenv("TMDB_READ_TOKEN")
	tmdbHttpClient, err := tmdbhttp.NewTmdbHttpClient(
		apiKey,
	)
	if err != nil {
		return err
	}
	tmdbService, err := tmdbhttp.NewTmdbService(tmdbHttpClient)
	if err != nil {
		return err
	}
	fileRenamer := &renamer.FileRenamer{}
	dirPathService, err := dirpath.NewDirPath(
		dirpath.TargetTvPathOption(moviePath),
	)
	if err != nil {
		return err
	}
	smlMth, err := singlemoviefile.NewSingleMovieFileMatcher(
		tmdbService,
		fileRenamer,
		moviePath,
	)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("singlemoviefile", smlMth)
	if err != nil {
		return err
	}
	aeMth, err := aniepfile.NewAnimeEpisodeFileMatcher(
		tmdbService,
		fileRenamer,
		tvPath,
	)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("aniepfile", aeMth)
	if err != nil {
		return err
	}
	ytsMth, err := ytsmovie.NewYTSMovieDirMatcher(
		tmdbHttpClient,
		fileRenamer,
		moviePath,
	)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("ytsmovie", ytsMth)
	if err != nil {
		return err
	}
	nekomeoMth, err := nekomoe.NewNeomoeMatcher(tmdbHttpClient, fileRenamer, tvPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("nekomeo", nekomeoMth)
	if err != nil {
		return err
	}
	// optional regex matcher
	regexMth, err := preregex.NewPreRegexMatcher("", tmdbService, fileRenamer, tvPath)
	if err == nil {
		err = matcherMgr.AddMatcher("regex", regexMth)
		if err != nil {
			return err
		}
	}
	megustaMth, err := megusta.NewMegustaMatcher(tmdbHttpClient, fileRenamer, tvPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("megusta", megustaMth)
	if err != nil {
		return err
	}
	ddhdtv, err := ddhdtv.NewDDHDTVMatcher(tmdbHttpClient, fileRenamer, tvPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("ddhdtv", ddhdtv)
	if err != nil {
		return err
	}
	movieShort, err := shorts.NewShortMatcher(tmdbHttpClient, fileRenamer, moviePath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("short", movieShort)
	if err != nil {
		return err
	}
	sevenAcgMth, err := sevenacg.NewSevenAcgMatcher(tmdbHttpClient, fileRenamer, tvPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("sevenacg", sevenAcgMth)
	if err != nil {
		return err
	}
	javFcMth, err := javfc.NewJavFCMatcher(fileRenamer, javPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("javfc", javFcMth)
	if err != nil {
		return err
	}
	huanyue, err := huanyue.NewHuanYueMatcher(tmdbHttpClient, fileRenamer, tvPath)
	if err != nil {
		return err
	}
	err = matcherMgr.AddMatcher("huanyue", huanyue)
	if err != nil {
		return err
	}
	tvepfile, err := tvepfile.NewTvEpisodeFileMatcher(".", tmdbService, fileRenamer, dirPathService)
	if err == nil {
		err = matcherMgr.AddMatcher("tvepfile", tvepfile)
		if err != nil {
			return err
		}
	}
	return nil
}
