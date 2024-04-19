package common

type MediaType int32

const (
	MediaTypeTrash MediaType = iota
	MediaTypeMovie
	MediaTypeTv
)

const (
	ValidStartYear = 1900
)

var (
	DefaultTmdbSearchOpts = map[string]string{
		"include_adult": "true",
	}
)
