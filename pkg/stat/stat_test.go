package stat

import (
	"asmediamgr/pkg/config"
	"testing"
)

func TestStatMovieEntry(t *testing.T) {
	input := &config.Configuration{
		StatDirs: []config.StatDir{
			{
				DirPath:   "./testdata",
				MediaType: "movie",
			},
		},
	}
	conf = input
	go statTask()
	cnt := 0
	for stat := range statChan {
		t.Logf("stat: %+v", stat)
		if stat.lasterr != nil {
			t.Errorf("failed to stat movie entry: %v", stat.lasterr)
		}
		if stat.msgType == msgTypeStat {
			cnt++
			if stat.entry == nil {
				t.Errorf("expected non-nil entry")
			}
			if stat.mediaTp != mediaTypeMovie {
				t.Errorf("expected movie media type")
			}
			if stat.tmdbid != 1234567 {
				t.Errorf("expected tmdbid 1234567")
			}
			if stat.movieStat == nil {
				t.Errorf("expected non-nil movieStat")
			}
			if stat.movieStat.movieFileNum != 1 {
				t.Errorf("expected 1 movie file")
			}
			if stat.movieStat.subtitleFileNum != 2 {
				t.Errorf("expected 2 subtitle files")
			}
		}
		if stat.msgType == msgTypeEnd {
			if cnt != 1 {
				t.Errorf("expected 1 stat message")
			}
			break
		}
	}
}
