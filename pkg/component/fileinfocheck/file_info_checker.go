package fileinfocheck

import (
	"asmediamgr/pkg/common"
	"fmt"
)

func CheckFileInfo(info *common.Info, checkers ...CheckerOption) error {
	for _, checker := range checkers {
		err := checker(info)
		if err != nil {
			return err
		}
	}
	return nil
}

type FileInfoChecker struct {
	info *common.Info
}

func NewFileInfoChecker(info *common.Info) *FileInfoChecker {
	checker := &FileInfoChecker{info: info}
	return checker
}

type CheckerOption func(info *common.Info) error

func CheckFileInfoHasOnlyOneFile() CheckerOption {
	return func(info *common.Info) error {
		if len(info.Subs) != 1 {
			return fmt.Errorf("len not 1, len:%d", len(info.Subs))
		}
		return nil
	}
}

func CheckFileInfoHasOnlyOneMediaFile() CheckerOption {
	return func(info *common.Info) error {
		if len(info.Subs) != 1 {
			return fmt.Errorf("len not 1, len:%d", len(info.Subs))
		}
		subInfo := info.Subs[0]
		if !common.IsMediaFile(subInfo.Ext) {
			return fmt.Errorf("not media file, ext:%s", subInfo.Ext)
		}
		return nil
	}
}

func CheckFileInfoHasOnlyOneFileSizeGreaterThan(size int64) CheckerOption {
	return func(info *common.Info) error {
		if len(info.Subs) != 1 {
			return fmt.Errorf("len not 1, len:%d", len(info.Subs))
		}
		subInfo := info.Subs[0]
		if subInfo.Size <= size {
			return fmt.Errorf("size not greater than %d, size:%d", size, subInfo.Size)
		}
		return nil
	}
}
