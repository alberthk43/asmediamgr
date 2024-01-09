package prometric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	curMontherDirsGuage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cur_mother_dirs",
		Help: "The current number of mother dirs",
	})
)

func CurMontherDirsSet(val float64) {
	curMontherDirsGuage.Set(val)
}

func CurMontherDirAdd() {
	curMontherDirsGuage.Inc()
}

func CurMontherDirDec() {
	curMontherDirsGuage.Dec()
}
