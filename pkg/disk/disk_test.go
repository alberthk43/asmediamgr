package disk

import (
	"testing"

	"github.com/go-kit/log"
)

func TestNewDisk(t *testing.T) {
	opts := &DiskServiceOpts{DryRunModeOpen: true}
	d, err := NewDiskService(opts)
	if err != nil {
		t.Errorf("NewDisk() error = %v", err)
	}
	if d == nil {
		t.Errorf("NewDisk() = %v, want not nil", d)
	}
}

func TestBuildTvEpTask(t *testing.T) {
	tests := []struct {
		name          string
		tvEpTask      *TvEpisodeRenameTask
		wantSeasonDir string
		wantEpFile    string
	}{
		{
			name: "normal",
			tvEpTask: &TvEpisodeRenameTask{
				OldPath:      "path/to/oldfile.ext",
				NewMotherDir: "path/to/mediabank",
				OriginalName: "original name",
				Year:         2021,
				Tmdbid:       123456789,
				Season:       2,
				Episode:      3,
			},
			wantSeasonDir: "path/to/mediabank/original name (2021) [tmdbid-123456789]/Season 2",
			wantEpFile:    "path/to/mediabank/original name (2021) [tmdbid-123456789]/Season 2/S02E03.ext",
		},
		{
			name: "escape",
			tvEpTask: &TvEpisodeRenameTask{
				OldPath:      "path/to/oldfile.ext",
				NewMotherDir: "path/to/mediabank",
				OriginalName: "name1 :?<> name2",
				Year:         2021,
				Tmdbid:       123456789,
				Season:       2,
				Episode:      3,
			},
			wantSeasonDir: "path/to/mediabank/name1      name2 (2021) [tmdbid-123456789]/Season 2",
			wantEpFile:    "path/to/mediabank/name1      name2 (2021) [tmdbid-123456789]/Season 2/S02E03.ext",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, ep, err := BuildNewEpisodePath(tt.tvEpTask)
			if err != nil {
				t.Fatalf("BuildNewEpisodePath() error = %v", err)
			}
			if dir != tt.wantSeasonDir {
				t.Errorf("BuildNewEpisodePath() \nreal %v, \nwant %v", dir, tt.wantSeasonDir)
			}
			if ep != tt.wantEpFile {
				t.Errorf("BuildNewEpisodePath() \nreal %v, \nwant %v", ep, tt.wantEpFile)
			}
		})
	}
}

func TestRenameTvEpFile(t *testing.T) {
	opts := &DiskServiceOpts{DryRunModeOpen: true, Logger: log.NewNopLogger()}
	d, err := NewDiskService(opts)
	if err != nil {
		t.Fatalf("NewDisk() error = %v", err)
	}
	task := &TvEpisodeRenameTask{
		OldPath:      "./disk.go",
		NewMotherDir: ".",
		OriginalName: "original name",
		Year:         2021,
		Tmdbid:       123456789,
		Season:       4,
		Episode:      5,
	}
	err = d.RenameTvEpisode(task)
	if err != nil {
		t.Errorf("RenameTvEpisode() error = %v", err)
	}
}
