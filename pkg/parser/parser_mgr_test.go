package parser

import (
	"testing"

	"github.com/go-kit/log"

	"asmediamgr/pkg/dirinfo"
)

type MockParser struct{}

func (p *MockParser) IsDefaultEnable() bool {
	return true
}

func (p *MockParser) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	return 0, nil
}

func (p *MockParser) Parse(entry *dirinfo.Entry, opts *ParserMgrRunOpts) (ok bool, err error) {
	return true, nil
}

func TestNewParserMgr(t *testing.T) {
	RegisterParser("parser1", &MockParser{})
	RegisterParser("parser2", &MockParser{})
	RegisterParser("parser3", &MockParser{})
	opts := &ParserMgrOpts{
		Logger:         nil,
		ConfigDir:      "config",
		EnableParsers:  []string{"parser1", "parser2"},
		DisableParsers: []string{"parser3", "parser2"},
	}
	pm, err := NewParserMgr(opts)
	if err != nil {
		t.Fatalf("NewParserMgr() error = %v", err)
	}
	if pm == nil {
		t.Fatalf("NewParserMgr() pm is nil")
	}
	if len(pm.parsers) != 1 {
		t.Fatalf("NewParserMgr() pm.parsers length = %v", len(pm.parsers))
	}
	if pm.parsers[0].parser == nil {
		t.Fatalf("NewParserMgr() pm.parsers[0] is nil")
	}
}

func TestPusnishAddTime(t *testing.T) {
	if punishAddTime(0).Minutes() != 0 {
		t.Errorf("punishAddTime(0) = %v", punishAddTime(0))
	}
	if punishAddTime(1).Minutes() != 1 {
		t.Errorf("punishAddTime(1) = %v", punishAddTime(1))
	}
	if punishAddTime(2).Minutes() != 2 {
		t.Errorf("punishAddTime(2) = %v", punishAddTime(2))
	}
	if punishAddTime(17).Minutes() != 65536 {
		t.Errorf("punishAddTime(17) = %v", punishAddTime(17))
	}
	if punishAddTime(200).Minutes() != 65536 {
		t.Errorf("punishAddTime(200) = %v", punishAddTime(200))
	}
}
