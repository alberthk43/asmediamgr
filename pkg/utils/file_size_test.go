package utils

import "testing"

func TestSizeStringToBytesNum(t *testing.T) {
	tests := []struct {
		str string
		num int64
	}{
		{"1", 1},
		{"1B", 1},
		{"1k", 1024},
		{"2K", 2048},
		{"1m", 1024 * 1024},
		{"3M", 1024 * 1024 * 3},
		{"1g", 1024 * 1024 * 1024},
		{"1024g", 1024 * 1024 * 1024 * 1024},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			num, err := SizeStringToBytesNum(tt.str)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if num != tt.num {
				t.Errorf("got: %d, want: %d", num, tt.num)
			}
		})
	}
}
