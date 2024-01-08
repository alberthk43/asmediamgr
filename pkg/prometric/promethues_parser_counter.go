package prometric

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	opsLoopMontherDirCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loop_mother_dir_ops_total",
		Help: "The total number of loop mother dir operations",
	})

	opsEntryCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "entry_ops_total",
		Help: "The total number of entry operations",
	})

	opsParserCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "parser_ops_total",
		Help: "The total number of parser operations",
	})

	succOpsParserCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "parser_succ_ops_total",
		Help: "The total number of parser operations",
	})

	muOpsCounterMapOfParsers sync.Mutex
	opsCounterMapOfParsers   = make(map[string]prometheus.Counter)
)

func LoopMontherDirInc() {
	opsLoopMontherDirCounter.Inc()
}

func EntryInc() {
	opsEntryCounter.Inc()
}

func ParserInc() {
	opsParserCounter.Inc()
}

func ParserSuccInc() {
	succOpsParserCounter.Inc()
}

func TemplateParserInc(name string) {
	muOpsCounterMapOfParsers.Lock()
	defer muOpsCounterMapOfParsers.Unlock()
	c, ok := opsCounterMapOfParsers[name]
	if !ok {
		c = promauto.NewCounter(prometheus.CounterOpts{
			Name: "parser_" + name + "_ops_total",
			Help: "The total number of parser operations",
		})
		opsCounterMapOfParsers[name] = c
	}
	c.Inc()
}
