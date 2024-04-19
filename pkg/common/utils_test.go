package common

import (
	"testing"
)

func TestParserTmdbDateStr(t *testing.T) {
	tests := []struct {
		datetimeStr string
		expected    *DateTime
	}{
		{
			datetimeStr: "2020-01-01",
			expected: &DateTime{
				Year:        2020,
				MonthOfYear: 1,
				DayOfMonth:  1,
			},
		},
		{
			datetimeStr: "2024-11-10",
			expected: &DateTime{
				Year:        2024,
				MonthOfYear: 11,
				DayOfMonth:  10,
			},
		},
	}
	for _, tt := range tests {
		dt, err := ParseTmdbDateStr(tt.datetimeStr)
		if err != nil {
			t.Fatalf("ParseTmdbDatetimeStr() datetimeStr = %s,  error = %v", tt.datetimeStr, err)
		}
		if dt == nil {
			t.Fatalf("ParseTmdbDatetimeStr() datetimeStr = %s, got = nil, expected = %v", tt.datetimeStr, tt.expected)
		}
		if dt.Year != tt.expected.Year || dt.MonthOfYear != tt.expected.MonthOfYear || dt.DayOfMonth != tt.expected.DayOfMonth {
			t.Fatalf("ParseTmdbDatetimeStr() datetimeStr = %s, got = %v, expected = %v", tt.datetimeStr, dt, tt.expected)
		}
	}
}

func TestChineseToNum(t *testing.T) {
	tests := []struct {
		chnStr   string
		expected int
		ok       bool
	}{
		{
			chnStr:   "零",
			expected: 0,
			ok:       true,
		},
		{
			chnStr:   "一",
			expected: 1,
			ok:       true,
		},
		{
			chnStr:   "二",
			expected: 2,
			ok:       true,
		},
		{
			chnStr:   "三",
			expected: 3,
			ok:       true,
		},
		{
			chnStr:   "四",
			expected: 4,
			ok:       true,
		},
		{
			chnStr:   "五",
			expected: 5,
			ok:       true,
		},
		{
			chnStr:   "六",
			expected: 6,
			ok:       true,
		},
		{
			chnStr:   "七",
			expected: 7,
			ok:       true,
		},
		{
			chnStr:   "八",
			expected: 8,
			ok:       true,
		},
		{
			chnStr:   "九",
			expected: 9,
			ok:       true,
		},
	}
	for _, tt := range tests {
		num, ok := ChineseToNum(tt.chnStr)
		if num != tt.expected || ok != tt.ok {
			t.Fatalf("ChineseToNum() chnStr = %s, got = %d, %t, expected = %d, %t", tt.chnStr, num, ok, tt.expected, tt.ok)
		}
	}
}
