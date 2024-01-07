package prometric

import (
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
