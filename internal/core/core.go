package core

import (
	"asmediamgr/internal/matcherservice"
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/matcher"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxErrTime  = 3
	sleepDump   = 60 * time.Second
	runInterval = 1 * time.Second
)

// TODO need refactor to run each single one
func Run(
	done <-chan interface{},
	path string,
) error {
	var err error
	ticker := time.NewTicker(runInterval)
	i := -1
	errTimes := make(map[string]int)
	var infoSlice []common.Info
	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			if i < 0 {
				infoSlice, err = dumpFullDir(path)
				if err != nil {
					return err
				}
				i = 0
			}
			if i < len(infoSlice) {
				info := infoSlice[i]
				i++
				if len(info.Subs) == 0 {
					return fmt.Errorf("no subs, dump err")
				}
				name := info.Subs[0].Name
				if n, ok := errTimes[name]; ok && n >= maxErrTime {
					continue
				}
				errTimes[name]++
				run(&info)
			} else {
				i = -1
				time.Sleep(sleepDump)
			}
		}
	}
}

func run(
	info *common.Info,
) {
	matchers := matcherservice.GetMatcherMgr().GetAllMatchers()
	log.Printf("** info:%s \n", info.DirPath)
	for _, sub := range info.Subs {
		log.Printf("**** name:%s%s dir:%t \n", sub.Name, sub.Ext, sub.IsDir)
		for _, path := range sub.Paths {
			log.Printf("******** path:%s\n", path)
		}
	}
	log.Printf("** info end\n")
	ok, err := matcher.Match(info, matchers)
	if err != nil {
		log.Printf("match result ok:%t sub[0]:%s%s err:%s\n", ok, info.Subs[0].Name, info.Subs[0].Ext, err)
		log.Printf("\n")
	}
}

func dumpFullDir(
	motherPath string,
) ([]common.Info, error) {
	stat, err := os.Stat(motherPath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not dir, path:%s", motherPath)
	}
	dir, err := os.Open(motherPath)
	if err != nil {
		return nil, err
	}
	subs, err := dir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	var infoSlice []common.Info
	for _, sub := range subs {
		info := common.Info{
			DirPath: motherPath,
		}
		if sub.IsDir() {
			subPath := filepath.Join(motherPath, sub.Name())
			err = filepath.Walk(subPath, func(filePath string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if filePath != subPath && f.IsDir() {
					return nil
				}
				if f.IsDir() {
					sub := common.Single{
						Name:  f.Name(),
						Ext:   "",
						IsDir: true,
						Size:  0,
					}
					info.Subs = append(info.Subs, sub)
				} else {
					ext := filepath.Ext(filePath)
					sub := common.Single{
						Name:  strings.TrimSuffix(f.Name(), ext),
						Ext:   ext,
						IsDir: false,
						Size:  f.Size(),
					}
					relativePath, err := filepath.Rel(motherPath, filepath.Dir(filePath))
					if err != nil {
						return err
					}
					sub.Paths = append(sub.Paths, relativePath)
					info.Subs = append(info.Subs, sub)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			subFile, err := os.Stat(filepath.Join(motherPath, sub.Name()))
			if err != nil {
				return nil, err
			}
			ext := filepath.Ext(sub.Name())
			sub := common.Single{
				Paths: nil, // TODO
				Name:  strings.TrimSuffix(sub.Name(), ext),
				Ext:   ext,
				IsDir: false,
				Size:  subFile.Size(),
			}
			info.Subs = append(info.Subs, sub)
		}
		infoSlice = append(infoSlice, info)
	}
	return infoSlice, nil
}
