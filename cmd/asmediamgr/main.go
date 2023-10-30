package main

import (
	"asmediamgr/internal/core"
	"asmediamgr/internal/matcherservice"
	"flag"
	"log"
	"os"
	"sync"
)

var (
	dirPath   string
	moviePath string
	tvPath    string
	javPath   string
)

func init() {
	flag.StringVar(&dirPath, "dirpath", ".", "dir path for all")
	flag.StringVar(&moviePath, "moviepath", ".", "target movie path")
	flag.StringVar(&tvPath, "tvpath", ".", "target tv path")
	flag.StringVar(&javPath, "javpath", ".", "target tv path")
}

func main() {
	flag.Parse()
	if err := doWork(); err != nil {
		log.Panic(err)
	}
}

func doWork() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := matcherservice.InitMatcher(moviePath, tvPath, javPath)
	if err != nil {
		return err
	}
	done := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Printf("stopped\n")
		err = core.Run(done, dirPath)
		if err != nil {
			log.Printf("run err:%s\n", err.Error())
		}
	}()
	sigs := make(chan os.Signal, 1)
	<-sigs
	close(done)
	wg.Wait()
	return nil
}
