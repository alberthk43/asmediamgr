package matcher

import (
	"asmediamgr/pkg/common"
	"fmt"
	"log"
	"time"
)

const (
	timeSleepBetweenTwoMatch = time.Millisecond * 399
)

type Matcher interface {
	Match(info *common.Info) (bool, error)
}

type FatalMatchErr interface {
	IsFatalMatchErr()
}

type MatcherMgr struct {
	matchers map[string]Matcher
}

func NewMatchMgr() *MatcherMgr {
	return &MatcherMgr{
		matchers: make(map[string]Matcher),
	}
}

func (mthMgr *MatcherMgr) AddMatcher(
	name string,
	matcher Matcher,
) error {
	if name == "" {
		return fmt.Errorf("name empty")
	}
	if matcher == nil {
		return fmt.Errorf("matcher nil")
	}
	if mthMgr.matchers == nil {
		mthMgr.matchers = make(map[string]Matcher)
	}
	mthMgr.matchers[name] = matcher
	return nil
}

func (mthMgr *MatcherMgr) GetAllMatchers() map[string]Matcher {
	return mthMgr.matchers
}

func Match(
	info *common.Info,
	matchers map[string]Matcher,
) (bool, error) {
	if info == nil {
		return false, fmt.Errorf("info nil")
	}
	for name, mth := range matchers {
		thisInfo := *info
		matched, err := mth.Match(&thisInfo)
		if err != nil {
			if _, ok := err.(FatalMatchErr); ok {
				log.Printf("match:%s err:%s fatal, stop current info, try next info\n", name, err)
				return false, err
			}
			log.Printf("match:%s err:%s, try next matcher\n", name, err)
			time.Sleep(timeSleepBetweenTwoMatch)
			continue
		} else {
			if matched {
				log.Printf("match:%s matched\n", name)
				time.Sleep(timeSleepBetweenTwoMatch)
				return matched, nil
			} else {
				log.Printf("match:%s not matched\n", name)
				time.Sleep(timeSleepBetweenTwoMatch)
				continue
			}
		}
	}
	return false, fmt.Errorf("no match")
}
