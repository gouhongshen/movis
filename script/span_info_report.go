package script

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"io"
	"log"
	_type "movis/type"
	"os"
	"time"
)

func paperReportForSpanInfo(tt OpType) {
	file, err := os.Create(_type.SpanReportDir + "report_" + time.Now().String())
	if err != nil {
		log.Panicf(err.Error())
	}

	for idx := range spanObjReqHeatmapData {
		report(file,
			spanObjReqHeatmapData[idx].Labels,
			spanObjReqHeatmapData[idx].Values,
			spanObjReqHeatmapData[idx].Title,
			func(s string) bool {
				if len(s) == 0 {
					return false
				}
				return true
			})
	}

	for idx := range spanObjReqThroughTimeData {
		report(file,
			spanObjReqThroughTimeData[idx].Labels,
			spanObjReqThroughTimeData[idx].Values,
			spanObjReqThroughTimeData[idx].Title,
			func(s string) bool {
				if len(s) < 1 {
					return false
				}
				return true
			})
	}

	for idx := range spanObjReadLatencyData {
		report(file,
			spanObjReadLatencyData[idx].Labels,
			spanObjReadLatencyData[idx].Values,
			spanObjReadLatencyData[idx].Title,
			func(s string) bool {
				if len(s) < 1 {
					return false
				}
				return true
			})
	}

}

func getBar(v float64, sum float64) string {
	maxLen := float64(500)

	bar := ""
	l := int(v / sum * maxLen)
	for i := 0; i < l; i++ {
		bar += "*"
	}
	return bar
}

func report(w io.Writer,
	labels []string, values []float64, title string,
	accept func(string) bool) {
	// assumed that values has descending order
	w.Write([]byte(title + "\n\n"))

	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"label", "rate", "cnt/latency(ms)"})

	sum := float64(0)
	for _, v := range values {
		sum += v
	}

	for i := 0; i < len(labels); i++ {
		bar := getBar(values[i], sum)
		if !accept(bar) {
			continue
		}
		t.AppendRow([]interface{}{labels[i], bar, values[i]})
	}

	t.Render()

	w.Write([]byte("\n\n\n"))

}
