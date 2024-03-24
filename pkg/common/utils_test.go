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
