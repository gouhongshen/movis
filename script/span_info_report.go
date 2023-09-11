package script

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"io"
	"log"
	_type "movis/type"
	"os"
	"path/filepath"
	"time"
)

func paperReportForSpanInfo(tt OpType) {
	path := fmt.Sprintf("%sreport_%s.report", _type.SpanReportDir, time.Now().String())
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		panic(err.Error())
	}

	file, err := os.Create(path)
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
	t.AppendHeader(table.Row{"label", "rate", "cnt"})

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
