package fileinfo

import (
	"regexp"
	"testing"
)

func TestMatchName(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		regexp   string
		expected string
		expectOk bool
	}{
		{
			name:     "case1",
			raw:      "[Nekomoe kissaten][Watashi no Shiawase na Kekkon][01-12][1080p][JPSC] tv tmdbid-196944",
			regexp:   `\[Nekomoe kissaten\]\[(?P<name>.*?)\].*tv tmdbid-196944$`,
			expected: "Watashi no Shiawase na Kekkon",
			expectOk: true,
		},
		{
			name:     "case2",
			raw:      "[Nekomoe kissaten][Watashi no Shiawase na Kekkon][01-12][1080p][JPSC]",
			regexp:   `\[Nekomoe kissaten\]\[(?P<name>.*?)\].*`,
			expected: "Watashi no Shiawase na Kekkon",
			expectOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := MatchName(tt.raw, regexp.MustCompile(tt.regexp))
			if tt.expectOk {
				if err != nil {
					t.Fatal(err)
				}
				if tt.expected != name {
					t.Fatalf("expected:%s actual:%s\n", tt.expected, name)
				}
			} else {
				if err == nil {
					t.Fatal()
				}
			}
		})
	}
}

func TestMatchTvEp(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		defaultSeason int32
		epRegex       string
		seasonRegex   string
		expectOk      bool
		expectSeason  int32
		expectEpNum   int32
	}{
		{
			name:          "case1",
			raw:           "xxx s03e11",
			defaultSeason: -1,
			epRegex:       `.*s(?P<season>\d+)e(?P<ep>\d+).*`,
			seasonRegex:   "",
			expectOk:      true,
			expectSeason:  3,
			expectEpNum:   11,
		},
		{
			name:          "case1",
			raw:           "xxx EP12 yyy",
			defaultSeason: 1,
			epRegex:       `.*EP(?P<ep>\d+).*`,
			seasonRegex:   "",
			expectOk:      true,
			expectSeason:  1,
			expectEpNum:   12,
		},
		{
			name:          "case2",
			raw:           "[Nekomoe kissaten][Watashi no Shiawase na Kekkon][06][1080p][JPSC]",
			defaultSeason: 1,
			epRegex:       `.*\[(?P<ep>\d{2})\].*`,
			seasonRegex:   "",
			expectOk:      true,
			expectSeason:  1,
			expectEpNum:   6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, epNum, err := MatchTvEp(
				tt.raw,
				tt.defaultSeason,
				regexp.MustCompile(tt.epRegex),
				seasonRegexp(t, tt.seasonRegex))
			if tt.expectOk {
				if err != nil {
					t.Fatal(err)
				}
				if tt.expectSeason != season {
					t.Fatalf("expected:%d actual:%d\n", tt.expectSeason, season)
				}
				if tt.expectEpNum != epNum {
					t.Fatalf("expected:%d actual:%d\n", tt.expectEpNum, epNum)
				}
			} else {
				if err == nil {
					t.Fatal()
				}
			}
		})
	}
}

func seasonRegexp(t *testing.T, str string) *regexp.Regexp {
	t.Helper()
	if str == "" {
		return nil
	} else {
		return regexp.MustCompile(`(?P<season>\d+)`)
	}
}
