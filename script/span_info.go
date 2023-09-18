package script

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"html/template"
	"log"
	"math"
	"movis/html"
	_type "movis/type"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type OpType int

const (
	AllFSOperation     OpType = 0
	S3FSOperation      OpType = 1
	LocalFSOperation   OpType = 2
	MemCacheOperation  OpType = 3
	DiskCacheOperation OpType = 4
)

var opType2Name = map[OpType]string{
	AllFSOperation:     "allFSOperation",
	S3FSOperation:      "s3FSOperation",
	LocalFSOperation:   "localFSOperation",
	MemCacheOperation:  "memCacheOperation",
	DiskCacheOperation: "diskCacheOperation",
}

func (op *OpType) String() string {
	return opType2Name[*op]
}

var spanObjReqHeatmapData []html.SignalLinePageData
var spanObjReqThroughTimeData []html.SignalLinePageData
var spanObjReadLatencyData []html.SignalLinePageData
var spanObjReqSizeThroughTimeData []html.SignalLinePageData
var spanObjReqStackInfoData []html.SignalLinePageData

var renderData html.LinePageData

var NodeType = []string{"CN", "TN"}

type SpanVis struct {
	db        *gorm.DB
	spanNames []string
}

var spanVis *SpanVis

func spanVisInit() {
	defer func() {

		renderData.Data = make([]html.SignalLinePageData, 0)
		spanObjReadLatencyData = make([]html.SignalLinePageData, 0)
		spanObjReqHeatmapData = make([]html.SignalLinePageData, 0)
		spanObjReqThroughTimeData = make([]html.SignalLinePageData, 0)
		spanObjReqSizeThroughTimeData = make([]html.SignalLinePageData, 0)

		spanObjReqStackInfoData = make([]html.SignalLinePageData, 0)
	}()

	if spanVis == nil {
		spanVis = new(SpanVis)

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/system?charset=utf8mb4&parseTime=True&loc=Local",
			_type.SrcUsrName, _type.SrcPassword, _type.SrcHost, _type.SrcPort)

		var err error
		spanVis.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Panicf("open %s\n failed", dsn)
		}
	}

	spanVis.spanNames = make([]string, 0)
	spanVis.db.Table("span_info").Select("distinct(span_name)").Find(&spanVis.spanNames)
}

func (s *SpanVis) webReport(w http.ResponseWriter, tt OpType) {
	wd, _ := os.Getwd()
	tmpl, err := template.ParseFiles(wd + "/html/line.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.saveWebReport(tmpl, w, tt)

	if _type.DstPort == "" {
		return
	}

	if err := tmpl.Execute(w, renderData); err != nil {
		fmt.Println(err.Error())
	}

}

func (s *SpanVis) saveWebReport(tmpl *template.Template, w http.ResponseWriter, tt OpType) {
	path := fmt.Sprintf("%s%s_%d.html", _type.SpanReportDir, tt.String(), time.Now().UnixMilli())
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		fmt.Println(err.Error())
		return
	}

	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err = tmpl.Execute(file, renderData); err != nil {
		fmt.Println(err.Error())
	}
}

func (s *SpanVis) generateReport(w http.ResponseWriter, tt OpType) {
	// user hope to generate a paper report
	paperReportForSpanInfo(tt)
	s.webReport(w, tt)
}

func AnalysisSpanInfoWithoutHttp() {
	LocalFSOperationHandler(nil, nil)
	S3FSOperationHandler(nil, nil)
	MemCacheOperationHandler(nil, nil)
	DiskCacheOperationHandler(nil, nil)
}

func LocalFSOperationHandler(w http.ResponseWriter, req *http.Request) {
	spanVisInit()
	spanVis.visualize(LocalFSOperation)
	spanVis.generateReport(w, LocalFSOperation)
}

func S3FSOperationHandler(w http.ResponseWriter, req *http.Request) {
	spanVisInit()
	spanVis.visualize(S3FSOperation)
	spanVis.generateReport(w, S3FSOperation)
}

func MemCacheOperationHandler(w http.ResponseWriter, req *http.Request) {
	spanVisInit()
	spanVis.visualize(MemCacheOperation)
	spanVis.generateReport(w, MemCacheOperation)
}

func DiskCacheOperationHandler(w http.ResponseWriter, req *http.Request) {
	spanVisInit()
	spanVis.visualize(DiskCacheOperation)
	spanVis.generateReport(w, DiskCacheOperation)
}

func (s *SpanVis) visualize(tt OpType) {
	s.visualize_ObjReqHeatmap(tt)
	s.visualize_ObjReqThroughTime(tt, 10)
	s.visualize_ObjReqLatency(tt, 10)
	s.visualize_ObjReqSizeChanges(tt, 10)
	s.visualize_ObjReqStackInfo(tt)
	s.visualize_StatementSpent(tt)
}

func (s *SpanVis) visualize_StatementSpent(tt OpType) {
	ret := make(map[string]map[string][]struct {
		stm      string
		duration float64
	})
	for _, nt := range NodeType {
		var traceIds []string
		s.db.Table("span_info").Select("distinct(trace_id)").
			Where(fmt.Sprintf("span_kind='statement' and node_type='%s'", nt)).
			Limit(1000).Find(&traceIds)

		for _, id := range traceIds {
			var data []struct {
				SpanName string
				Duration float64
				Extra    string
			}

			s.db.Table("span_info").Select("span_name, duration, extra").
				Where(fmt.Sprintf("span_kind='statement' and trace_id='%s'", id)).
				Order("span_name").Find(&data)

			tmp := make(map[string]float64)
			stm := make(map[string]string)
			for i := range data {
				tmp[data[i].SpanName] += data[i].Duration
				stm[data[i].SpanName] = data[i].Extra
			}

			for k, _ := range tmp {
				if p := ret[nt]; p == nil {
					ret[nt] = make(map[string][]struct {
						stm      string
						duration float64
					})
				}
				ret[nt][k] = append(ret[nt][k], struct {
					stm      string
					duration float64
				}{stm[k], tmp[k] / (1000 * 1000)})
			}
		}
	}
	reportStatement(ret)
}

func (s *SpanVis) visualize_ObjReqStackInfo(tt OpType) {
	for _, nt := range NodeType {
		for _, nn := range s.spanNames {
			var data []struct {
				StackName string
				Count     float64
			}

			s.db.Table("span_info").
				Select("json_unquote(json_extract(`extra`, '$.stack')) as stack_name, " +
					"count(*) as count").
				Where(fmt.Sprintf("node_type='%s' and span_name='%s' and span_kind='%s'", nt, nn, tt.String())).
				Group("stack_name").
				Order("count desc").Find(&data)

			var labels []string
			var values []float64
			for i := range data {
				labels = append(labels, data[i].StackName)
				values = append(values, data[i].Count)
			}

			s.appendToTar(&spanObjReqStackInfoData, labels, values,
				"stack name", "numbers", "line",
				nt+"  "+tt.String()+"  "+nn+":  Obj Req stack info")
		}
	}
}

func (s *SpanVis) visualize_ObjReqSizeChanges(tt OpType, duration int) {
	for _, nt := range NodeType {
		for _, nn := range s.spanNames {
			var data []struct {
				Timestamp int64
				AvgSize   float64
			}

			s.db.Table("span_info").
				Select(fmt.Sprintf("floor(unix_timestamp(`start_time`)/%d)*%d as timestamp, "+
					"avg(cast(json_unquote(json_extract(`extra`, '$.size')) as decimal(10,2))) as avg_size",
					duration, duration)).
				Where(fmt.Sprintf("node_type='%s' and span_name='%s' and span_kind='%s'", nt, nn, tt.String())).
				Group("timestamp").
				Order("timestamp").Find(&data)

			var labels []string
			var values []float64
			for i := range data {
				labels = append(labels, time.Unix(data[i].Timestamp, 0).String())
				values = append(values, math.Floor(data[i].AvgSize/1024*100)/100)
			}

			s.appendToTar(&spanObjReqSizeThroughTimeData, labels, values,
				"timestamp", fmt.Sprintf("average size (KB) in each %ds", duration), "line",
				nt+"  "+tt.String()+"  "+nn+":  Obj Req average size")
		}
	}
	s.appendToRenderData(spanObjReqSizeThroughTimeData)
}

func (s *SpanVis) visualize_ObjReqLatency(tt OpType, duration int) {
	for _, nt := range NodeType {
		for _, nn := range s.spanNames {
			var data []struct {
				Timestamp   int64
				AvgDuration float64
			}

			s.db.Table("span_info").
				Select(fmt.Sprintf("floor(unix_timestamp(`start_time`)/%d)*%d as timestamp, "+
					"avg(`duration`) as avg_duration", duration, duration)).
				Where(fmt.Sprintf("node_type='%s' and span_name='%s' and span_kind='%s'", nt, nn, tt.String())).
				Group("timestamp").
				Order("timestamp").Find(&data)

			var labels []string
			var values []float64

			for i := range data {

				labels = append(labels, time.Unix(data[i].Timestamp, 0).String())
				values = append(values, data[i].AvgDuration/(1000*1000))
			}

			s.appendToTar(&spanObjReadLatencyData, labels, values, "timestamp",
				fmt.Sprintf("average latency (ms) in each %ds", duration), "line",
				nt+"  "+tt.String()+"  "+nn+":  Obj Req average latency")

		}
	}
	s.appendToRenderData(spanObjReadLatencyData)
}

func (s *SpanVis) visualize_ObjReqThroughTime(tt OpType, duration int) {
	for _, nt := range NodeType {
		for _, nn := range s.spanNames {
			var data []struct {
				Timestamp int64
				Count     float64
			}

			s.db.Table("span_info").
				Select(fmt.Sprintf("floor(unix_timestamp(`start_time`)/%d)*%d as timestamp, count(*) as count",
					duration, duration)).
				Where(fmt.Sprintf("node_type='%s' and span_kind='%s' and span_name='%s'", nt, tt.String(), nn)).
				Group("timestamp").
				Order("timestamp").Find(&data)

			var labels []string
			var values []float64

			for i := range data {
				labels = append(labels, time.Unix(data[i].Timestamp, 0).String())
				values = append(values, data[i].Count)
			}

			s.appendToTar(&spanObjReqThroughTimeData, labels, values,
				"timestamp", fmt.Sprintf("requested numbers in each %ds", duration), "line",
				nt+"  "+tt.String()+"  "+nn+":  Obj Req Through Time")
		}
	}
	s.appendToRenderData(spanObjReqThroughTimeData)
}

func (s *SpanVis) unmarshExtra(extra string) map[string]interface{} {
	var ret map[string]interface{}
	if err := json.Unmarshal([]byte(extra), &ret); err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
	}
	return ret
}

func (s *SpanVis) visualize_ObjReqHeatmap(tt OpType) {
	for _, nt := range NodeType {
		for _, nn := range s.spanNames {
			var data []struct {
				Name  string
				Count float64
			}

			s.db.Table("span_info").
				Where(fmt.Sprintf("node_type='%s' and span_kind='%s' and span_name='%s'", nt, tt.String(), nn)).
				Select("JSON_EXTRACT(extra, '$.name') AS name, COUNT(*) AS count").
				Group("name").Order("count desc").Find(&data)

			var labels []string
			var values []float64

			for i := range data {
				labels = append(labels, data[i].Name)
				values = append(values, data[i].Count)
			}

			s.appendToTar(&spanObjReqHeatmapData, labels, values,
				"object name", "object requested numbers", "line",
				nt+"  "+tt.String()+"  "+nn+":  Obj Request Heatmap")
		}
	}

	s.appendToRenderData(spanObjReqHeatmapData)
}

func (s *SpanVis) appendToRenderData(data []html.SignalLinePageData) {
	renderData.Data = append(renderData.Data, data...)
}

func (s *SpanVis) appendToTar(tar *[]html.SignalLinePageData,
	labels []string, values []float64, xaxis string, yaxis string,
	chartType string, title string) {

	if len(labels) == 0 {
		return
	}

	*tar = append(*tar, html.SignalLinePageData{
		Labels:    labels,
		Values:    values,
		XAxis:     xaxis,
		YAxis:     yaxis,
		ChartType: chartType,
		Title:     title,
	})
}
