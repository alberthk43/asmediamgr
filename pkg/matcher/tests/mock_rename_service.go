package tests

import (
	"asmediamgr/pkg/component/renamer"
	"fmt"
)

type MockRenamer struct {
	Expected []renamer.RenameRecord
}

func (r *MockRenamer) Rename(records []renamer.RenameRecord) error {
	for _, record := range records {
		if err := r.matchRecord(record); err != nil {
			return err
		}
	}
	return nil
}

func (r *MockRenamer) matchRecord(record renamer.RenameRecord) error {
	for i := 0; i < len(r.Expected); i++ {
		if err := isSamePath(record.Old, r.Expected[i].Old); err == nil {
			if err := isSamePath(record.New, r.Expected[i].New); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("not match, input:%v, expected in:%v", record, r.Expected)
}

func isSamePath(a, b renamer.Path) error {
	if len(a) != len(b) {
		return fmt.Errorf("len err, a:%v b:%v", a, b)
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return fmt.Errorf("path err, a:%v b:%v", a[i], b[i])
		}
	}
	return nil
}
