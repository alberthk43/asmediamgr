package common

type MediaType int32

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeMovie
	MediaTypeTv
)

const (
	ValidStartYear = 1900
)
