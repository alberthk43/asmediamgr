package utils

var (
	subtitleExt = map[string]struct{}{
		".srt": {},
		".ass": {},
	}
)

func IsSubtitleExt(ext string) bool {
	_, ok := subtitleExt[ext]
	return ok
}
