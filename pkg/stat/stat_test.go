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
		if stat.Lasterr != nil {
			t.Errorf("failed to stat movie entry: %v", stat.Lasterr)
		}
		if stat.MsgType == msgTypeStat {
			cnt++
			if stat.Entry == nil {
				t.Errorf("expected non-nil entry")
			}
			if stat.MediaTp != mediaTypeMovie {
				t.Errorf("expected movie media type")
			}
			if stat.Tmdbid != 1234567 {
				t.Errorf("expected tmdbid 1234567")
			}
			if stat.MovieStat == nil {
				t.Errorf("expected non-nil movieStat")
			}
			if stat.MovieStat.MovieFileNum != 1 {
				t.Errorf("expected 1 movie file")
			}
			if stat.MovieStat.SubtitleFileNum != 2 {
				t.Errorf("expected 2 subtitle files")
			}
		}
		if stat.MsgType == msgTypeEnd {
			if cnt != 1 {
				t.Errorf("expected 1 stat message")
			}
			break
		}
	}
}
