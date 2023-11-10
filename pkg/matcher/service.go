package matcher

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamer"
)

// TmdbServive define the tmdb serivice for search provided by tmdb website database
type TmdbService interface {
	// SearchTvByTmdbID search tv info by tmdb id, return nil if not found
	SearchTvByTmdbID(tmdbID int64) (*common.MatchedTV, error)

	// SearchTvByName search tv info by name, return nil if not found or too many results
	SearchTvByName(name string) (*common.MatchedTV, error)

	// SearchMovieByTmdbId
	SearchMovieByTmdbID(tmdbId int64) (*common.MatchedMovie, error)

	// SearchMovieByName
	SearchMovieByName(name string, year int) (*common.MatchedMovie, error)
}

// RenamerService define the rename service, commonly used at final stage of renaming files and directories
type RenamerService interface {
	// Rename to the rename
	Rename(records []renamer.RenameRecord) error
}

type TargetService interface {
	TargetDir() string
}
