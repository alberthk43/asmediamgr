package tests

import "asmediamgr/pkg/matcher"

var _ matcher.TargetService = (*MockTargetService)(nil)

type MockTargetService struct {
	TargetDirPath string
}

func (s *MockTargetService) TargetDir() string {
	return s.TargetDirPath
}
