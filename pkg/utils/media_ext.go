package utils

var (
	mediaExt = map[string]struct{}{
		".mp4":  {},
		".MP4":  {},
		".mkv":  {},
		".MKV":  {},
		".avi":  {},
		".AVI":  {},
		".rmvb": {},
		".RMVB": {},
		".rm":   {},
		".RM":   {},
		".wmv":  {},
		".WMV":  {},
		".flv":  {},
		".FLV":  {},
		".mov":  {},
		".MOV":  {},
		".m4v":  {},
		".M4V":  {},
	}
)

func IsMediaExt(ext string) bool {
	_, ok := mediaExt[ext]
	return ok
}
