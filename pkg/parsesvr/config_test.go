package parsesvr

import (
	"testing"
	"time"
)

var (
	testConfigFile = "tests/test_config.toml"
)

func TestLoadConfiguration(t *testing.T) {
	c, err := LoadConfigurationFromFile(testConfigFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	expect := &Configuration{
		ServiceConfDir: "path/to/serviceconf",
		ParserConfDir:  "path/to/parserconf",
		MotherDirs: []MontherDir{
			{
				DirPath:       "path/to/motherdir1",
				SleepInterval: time.Duration(1) * time.Hour,
			},
			{
				DirPath:       "path/to/motherdir2",
				SleepInterval: time.Duration(9)*time.Minute + time.Duration(13)*time.Second,
			},
		},
	}
	testConfigurationSame(t, expect, c)
}

func testConfigurationSame(t testing.TB, expect, real *Configuration) {
	if expect.ServiceConfDir != real.ServiceConfDir {
		t.Errorf("ServiceConfDir: expected %s, got %s", expect.ServiceConfDir, real.ServiceConfDir)
	}
	if expect.ParserConfDir != real.ParserConfDir {
		t.Errorf("ParserConfDir: expected %s, got %s", expect.ParserConfDir, real.ParserConfDir)
	}
	if len(expect.MotherDirs) != len(real.MotherDirs) {
		t.Fatalf("MotherDirs: expected %d, got %d", len(expect.MotherDirs), len(real.MotherDirs))
	}
	for i := range expect.MotherDirs {
		if expect.MotherDirs[i].DirPath != real.MotherDirs[i].DirPath {
			t.Errorf("MotherDirs[%d].DirPath: expected %s, got %s", i, expect.MotherDirs[i].DirPath, real.MotherDirs[i].DirPath)
		}
		if expect.MotherDirs[i].SleepInterval != real.MotherDirs[i].SleepInterval {
			t.Errorf("MotherDirs[%d].SleepInterval: expected %s, got %s", i, expect.MotherDirs[i].SleepInterval, real.MotherDirs[i].SleepInterval)
		}
	}
}
