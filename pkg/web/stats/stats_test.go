package stats

import "testing"

func TestParseMovieInfo(t *testing.T) {
	/// arrange
	tests := []struct {
		tname       string
		dirName     string
		expectMedia *Media
	}{
		{
			"normal",
			"XXX (2023) [tmdbid-1010]",
			&Media{
				MediaType:    Movie,
				TmdbID:       1010,
				OriginalName: "XXX",
				Year:         2023,
			},
		},
		{
			"badTmdbIdNaming",
			"XXX (2023) [tmdbid-ABCD]",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			m, err := parseMovieInfoFromName(tt.dirName)
			if tt.expectMedia == nil {
				if err == nil {
					t.Fatal("expect error, but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expect nil, but got %v", err)
				}
				if m.MediaType != tt.expectMedia.MediaType {
					t.Fatalf("expect %v, but got %v", tt.expectMedia.MediaType, m.MediaType)
				}
				if m.TmdbID != tt.expectMedia.TmdbID {
					t.Fatalf("expect %v, but got %v", tt.expectMedia.TmdbID, m.TmdbID)
				}
				if m.OriginalName != tt.expectMedia.OriginalName {
					t.Fatalf("expect %v, but got %v", tt.expectMedia.OriginalName, m.OriginalName)
				}
				if m.Year != tt.expectMedia.Year {
					t.Fatalf("expect %v, but got %v", tt.expectMedia.Year, m.Year)
				}
			}
		})
	}
}

func TestGatherStats(t *testing.T) {
	/// arrange
	moviePath := `./tests/dirs`
	ms, err := GatherMovieStats(moviePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms.Medias) != 1 {
		t.Fatal(len(ms.Medias))
	}
	movies, ok := ms.Medias[Movie]
	if !ok {
		t.Fatal("no movie")
	}
	if len(movies) != 2 {
		t.Fatal(len(movies))
	}
	movieMapping := make(map[int64]*Media)
	for i := 0; i < len(movies); i++ {
		m := &movies[i]
		movieMapping[m.TmdbID] = m
	}
	movie1010, ok := movieMapping[1010]
	if !ok {
		t.Fatal("no movie 1010")
	}
	if movie1010.MediaType != Movie {
		t.Fatal(movie1010.MediaType)
	}
	if movie1010.OriginalName != "XXX" {
		t.Fatal(movie1010.OriginalName)
	}
	if movie1010.Year != 2023 {
		t.Fatal(movie1010.Year)
	}
	movie2020, ok := movieMapping[2020]
	if !ok {
		t.Fatal("no movie 2020")
	}
	if movie2020.MediaType != Movie {
		t.Fatal(movie2020.MediaType)
	}
	if movie2020.OriginalName != "YYY" {
		t.Fatal(movie2020.OriginalName)
	}
	if movie2020.Year != 2029 {
		t.Fatal(movie2020.Year)
	}

	/// action
	/// assert
}
