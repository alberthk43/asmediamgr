package tests

import "asmediamgr/pkg/matcher"

var _ matcher.PathService = (*MockTargetService)(nil)

type MockTargetService struct {
	TargetPathPath string
	MotherPathPath string
	ConfigPathPath string
}

func (s *MockTargetService) TargetTvPath() string {
	return s.TargetPathPath
}

func (s *MockTargetService) MotherPath() string {
	return s.MotherPathPath
}

func (s *MockTargetService) ConfigPath() string {
	return s.ConfigPathPath
}
