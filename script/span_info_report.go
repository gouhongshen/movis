package script

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"io"
	"log"
	_type "movis/type"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func paperReportForSpanInfo(tt OpType) {
	path := fmt.Sprintf("%s%s_%d.report", _type.SpanReportDir, tt.String(), time.Now().UnixMilli())
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
			spanObjReqHeatmapData[idx].XAxis,
			spanObjReqHeatmapData[idx].YAxis,
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
			spanObjReqThroughTimeData[idx].XAxis,
			spanObjReqThroughTimeData[idx].YAxis,
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
			spanObjReadLatencyData[idx].XAxis,
			spanObjReadLatencyData[idx].YAxis,
			func(s string) bool {
				if len(s) < 1 {
					return false
				}
				return true
			})
	}

	for idx := range spanObjReqSizeThroughTimeData {
		report(file,
			spanObjReqSizeThroughTimeData[idx].Labels,
			spanObjReqSizeThroughTimeData[idx].Values,
			spanObjReqSizeThroughTimeData[idx].Title,
			spanObjReqSizeThroughTimeData[idx].XAxis,
			spanObjReqSizeThroughTimeData[idx].YAxis,
			func(s string) bool {
				if len(s) < 1 {
					return false
				}
				return true
			})
	}

	for idx := range spanObjReqStackInfoData {
		report(file,
			spanObjReqStackInfoData[idx].Labels,
			spanObjReqStackInfoData[idx].Values,
			spanObjReqStackInfoData[idx].Title,
			spanObjReqStackInfoData[idx].XAxis,
			spanObjReqStackInfoData[idx].YAxis,
			func(s string) bool {
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
	xaxis string, yaxis string,
	accept func(string) bool) {
	// assumed that values has descending order
	w.Write([]byte(title + "\n" + "x-axis: " + xaxis + "; y-axis: " + yaxis + "\n\n"))

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

	if len(labels) > 0 {
		w.Write([]byte("\n"))

		t = table.NewWriter()
		t.SetOutputMirror(w)
		t.AppendHeader(table.Row{"label", "val"})

		t.AppendRow([]interface{}{"total", len(labels)})
		t.AppendRow([]interface{}{"average", sum / float64(len(labels))})

		nVals := make([]float64, len(values))
		copy(nVals, values)
		sort.Slice(nVals, func(i, j int) bool { return nVals[i] < nVals[j] })

		t.AppendRow([]interface{}{"median", nVals[len(nVals)/2]})
		t.AppendRow([]interface{}{"maximum", nVals[len(nVals)-1]})
		t.AppendRow([]interface{}{"minimum", nVals[0]})
		t.AppendRow([]interface{}{"95-percent", nVals[len(nVals)*95/100]})

		step := (nVals[len(nVals)-1] - nVals[0]) / 20
		for i, j := 0, 0; i < len(nVals); {
			cnt := 0
			for j = i; j < len(nVals); j++ {
				if nVals[i]+step >= nVals[j] {
					cnt++
				} else {
					break
				}
			}
			t.AppendRow([]interface{}{fmt.Sprintf("density_%02d", i),
				fmt.Sprintf("%6.3f%s: [%6.3f, %6.3f]", float64(cnt)/float64(len(nVals))*100, "%",
					nVals[i], nVals[j-1])})
			i = j
		}

		t.Render()
	}

	w.Write([]byte("\n\n\n"))

}
