package common

import (
	"fmt"
	"time"
)

type DateTime struct {
	Year        int
	MonthOfYear int
	DayOfMonth  int
}

// ParseTmdbDateStr parses tmdb date string to DateTime struct
func ParseTmdbDateStr(tmdbDateTimeStr string) (dt *DateTime, err error) {
	tm, err := time.Parse("2006-01-02", tmdbDateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("time.Parse() str: %s, error: %v", tmdbDateTimeStr, err)
	}
	dt = &DateTime{
		Year:        tm.Year(),
		MonthOfYear: int(tm.Month()),
		DayOfMonth:  tm.Day(),
	}
	return dt, nil
}

// ChineseToNum converts Chinese number to Arabic number
// Notes: only support 0-9
func ChineseToNum(chnStr string) (num int, ok bool) {
	switch chnStr {
	case "零":
		num = 0
	case "一":
		num = 1
	case "二":
		num = 2
	case "三":
		num = 3
	case "四":
		num = 4
	case "五":
		num = 5
	case "六":
		num = 6
	case "七":
		num = 7
	case "八":
		num = 8
	case "九":
		num = 9
	default:
		num = -1
	}
	return num, num >= 0
}
