package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alberthk43/asmediamgr/pkg/dirinfo"
)

func FileAtLeast(file *dirinfo.File, atLeastBytesNum int64) bool {
	return file.BytesNum >= atLeastBytesNum
}

const (
	mul = 1024
)

type sizeInfo struct {
	suffix string
	smul   int64
}

var suffixMapping = []sizeInfo{
	{"g", mul * mul * mul},
	{"m", mul * mul},
	{"k", mul},
	{"", 1},
}

func SizeStringToBytesNum(str string) (int64, error) {
	str, _ = strings.CutSuffix(str, "B")
	str = strings.ToLower(str)
	for _, info := range suffixMapping {
		numStr, ok := strings.CutSuffix(str, info.suffix)
		if ok {
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return num * info.smul, nil
		}
	}
	return 0, fmt.Errorf("invalid size string: %s", str)
}

func BytesNumToSizeString(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/1024/1024)
	}
	return fmt.Sprintf("%.2fGB", float64(size)/1024/1024/1024)
}
