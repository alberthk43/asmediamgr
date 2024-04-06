package utils

import (
	"asmediamgr/pkg/dirinfo"
	"fmt"
	"strconv"
	"strings"
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
