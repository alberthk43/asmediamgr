package server

import (
	"testing"
	"time"
)

func TestSortParserInfoSlice(t *testing.T) {
	parserInfoSlice := parserInfoSlice{
		parserInfo{priority: 2},
		parserInfo{priority: 1},
	}
	sorted := sortParserInfo(parserInfoSlice)
	if len(sorted) != 2 {
		t.Fatal("len(sorted) != 2")
	}
	if sorted[0].priority != 1 {
		t.Fatal("sorted[0].priority != 1")
	}
	if sorted[1].priority != 2 {
		t.Fatal("sorted[1].priority != 2")
	}
}

func TestNextRetryTime(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2019-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	next := nextRetryTime(3, now)
	if next.Sub(now) != time.Second*8 {
		t.Fatal(err)
	}
}
