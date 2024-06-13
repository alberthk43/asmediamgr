package common

type MediaType int32

const (
	MediaTypeTrash MediaType = iota
	MediaTypeMovie
	MediaTypeTv
)

var (
	DefaultTmdbSearchOpts = map[string]string{
		"include_adult": "true",
	}
)
