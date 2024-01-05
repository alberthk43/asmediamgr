package server

import "testing"

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
