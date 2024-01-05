package tvepfile

import "testing"

func TestLoadConfiguration(t *testing.T) {
	configPath := "./tests/aniteam,test.toml"
	config, err := loadConfiguration(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Predefined) != 2 {
		t.Fatalf("expect 2 predefined, got %d", len(config.Predefined))
	}
	if config.Predefined[0].Name != "ABC" {
		t.Fatalf("expect name is ABC, got %s", config.Predefined[0].Name)
	}
	if config.Predefined[0].TmdbId != 123 {
		t.Fatalf("expect tmdbid is 123, got %d", config.Predefined[0].TmdbId)
	}
	if config.Predefined[0].SeasonNum != 2 {
		t.Fatalf("expect season_num is 2, got %d", config.Predefined[0].SeasonNum)
	}
	if config.Predefined[1].Name != "DEF" {
		t.Fatalf("expect name is DEF, got %s", config.Predefined[1].Name)
	}
	if config.Predefined[1].TmdbId != 456 {
		t.Fatalf("expect tmdbid is 456, got %d", config.Predefined[1].TmdbId)
	}
	if config.Predefined[1].SeasonNum != 99 {
		t.Fatalf("expect season_num is 99, got %d", config.Predefined[1].SeasonNum)
	}
}
