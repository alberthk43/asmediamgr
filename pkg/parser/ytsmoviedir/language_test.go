package ytsmoviedir

import "testing"

func TestStdLanguageMatch(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "eng", want: "en"},
		{name: "ara", want: "ar"},
		{name: "baq", want: "es"},
		{name: "cat", want: "ca"},
		{name: "cze", want: "cs"},
		{name: "dan", want: "da"},
		{name: "dut", want: "nl"},
		{name: "fil", want: "fil"},
		{name: "fin", want: "fi"},
		{name: "fre", want: "fr"},
		{name: "ger", want: "de"},
		{name: "glg", want: "es"},
		{name: "gre", want: "el"},
		{name: "heb", want: "he"},
		{name: "hrv", want: "hr"},
		{name: "hun", want: "hu"},
		{name: "ind", want: "id"},
		{name: "ita", want: "it"},
		{name: "jpn", want: "ja"},
		{name: "kor", want: "ko"},
		{name: "may", want: "ms"},
		{name: "nob", want: "da"},
		{name: "pol", want: "pl"},
		{name: "por", want: "pt"},
		{name: "rum", want: "ro"},
		{name: "rus", want: "ru"},
		{name: "spa", want: "es"},
		{name: "swe", want: "sv"},
		{name: "tha", want: "th"},
		{name: "tur", want: "tr"},
		{name: "ukr", want: "uk"},
		{name: "vie", want: "vi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLanguageFromFileName(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("getLanguageFromFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpecialLanguageMatch(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "English", want: "en"},
		{name: "Simplified.chi", want: "zh-Hans"},
		{name: "Traditional.chi", want: "zh-Hant"},
		{name: "Brazilian.por", want: "pt-BR"},
		{name: "Latin American.spa", want: "es-419"},
		{name: "SDH.eng.HI", want: "en.sdh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLanguageFromFileName(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("getLanguageFromFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
