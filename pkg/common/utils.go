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
