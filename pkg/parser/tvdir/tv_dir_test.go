package tvdir

import (
	"os"
	"testing"

	"github.com/go-kit/log"
)

func TestTvDirTmdbidAndSxExLikePattern(t *testing.T) {
	cfg := `
[[patterns]]
dir_pattern = ".* tv tmdbid-(?P<tmdbid>\\d+)$"
episode_file_at_least = "10MB"
episode_pattern = "[Ss](?P<season>\\d+)[Ee](?P<episode>\\d+)"
`
	if err := os.WriteFile(`tmp.cfg`, []byte(cfg), os.FileMode(0755)); err != nil {
		t.Fatal(err)
	}
	p := &TvDir{}
	if _, err := p.Init(`tmp.cfg`, log.NewNopLogger()); err != nil {
		t.Fatal(err)
	}
}
