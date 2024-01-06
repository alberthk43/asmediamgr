package utils

var (
	mediaExt = map[string]struct{}{
		".mp4":  {},
		".mkv":  {},
		".avi":  {},
		".rmvb": {},
		".rm":   {},
		".wmv":  {},
		".flv":  {},
		".mov":  {},
		".mpg":  {},
		".mpeg": {},
		".m4v":  {},
	}
)

func IsMediaExt(ext string) bool {
	_, ok := mediaExt[ext]
	return ok
}
