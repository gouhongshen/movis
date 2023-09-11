package script

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"html/template"
	"log"
	"movis/html"
	_type "movis/type"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type OpType int

const (
	AllFSOperation   OpType = 0
	S3FSOperation    OpType = 1
	LocalFSOperation OpType = 2
)

var opType2Name = map[OpType]string{
	AllFSOperation:   "allFSOperation",
	S3FSOperation:    "s3FSOperation",
	LocalFSOperation: "localFSOperation",
}

func (op *OpType) String() string {
	return opType2Name[*op]
}

var spanObjReqHeatmapData []html.SignalLinePageData
var spanObjReqThroughTimeData []html.SignalLinePageData
var spanObjReadLatencyData []html.SignalLinePageData

var renderData html.LinePageData

type SpanVis struct {
	db           *gorm.DB
	infos        []_type.SpanInfoTable
	readRecords  []_type.SpanInfoTable
	writeRecords []_type.SpanInfoTable
}

var spanVis *SpanVis

func spanVisInit() {
	defer func() {
		spanVis.infos = make([]_type.SpanInfoTable, 0)
		spanVis.readRecords = make([]_type.SpanInfoTable, 0)
		spanVis.writeRecords = make([]_type.SpanInfoTable, 0)

		renderData.Data = make([]html.SignalLinePageData, 0)
		spanObjReadLatencyData = make([]html.SignalLinePageData, 0)
		spanObjReqHeatmapData = make([]html.SignalLinePageData, 0)
		spanObjReqThroughTimeData = make([]html.SignalLinePageData, 0)
	}()

	spanVis = new(SpanVis)

	if _type.SourceFile != "" {
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/system?charset=utf8mb4&parseTime=True&loc=Local",
		_type.SrcUsrName, _type.SrcPassword, _type.SrcHost, _type.SrcPort)

	var err error
	spanVis.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("open %s\n failed", dsn)
	}
}

func (s *SpanVis) webReport(w http.ResponseWriter) {
	wd, _ := os.Getwd()
	tmpl, err := template.ParseFiles(wd + "/html/line.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, renderData); err != nil {
		fmt.Println(err.Error())
	}
}

func (s *SpanVis) generateReport(w http.ResponseWriter, tt OpType) {
	// user hope to generate a paper report
	paperReportForSpanInfo(tt)

	if _type.DstPort != "" {
		s.webReport(w)
	}
}

func AnalysisSpanInfoWithoutHttp() {
	LocalFSOperationHandler(nil, nil)
	S3FSOperationHandler(nil, nil)
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

func (s *SpanVis) decodeCSV(tt OpType) {
	file, err := os.Open(_type.SourceFile)
	if err != nil {
		panic(err.Error())
	}

	funcs := _type.SpanInfoMemberSetFunc()
	reader := csv.NewReader(file)
	heads, err := reader.Read()
	for {
		records, err := reader.Read()
		if err != nil {
			return
		}
		info := _type.SpanInfoTable{}
		for i := 0; i < len(heads); i++ {
			vals := _type.SpanInfoMapTag2Values(&info)
			vals[heads[i]].Set(funcs[heads[i]](records[i]))
		}

		if info.SpanKind != tt.String() {
			continue
		}
		s.infos = append(s.infos, info)
	}

}

func (s *SpanVis) saveCSV(tt OpType) {
	name := fmt.Sprintf("./src_data/%s_%s.csv", tt.String(), time.Now().String())
	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		panic(err.Error())
	}

	file, err := os.Create(name)
	if err != nil {
		log.Panic(err.Error())
		return
	}
	writer := csv.NewWriter(file)
	writer.UseCRLF = false

	if err = writer.Write(_type.SpanInfoTableCSVHead()); err != nil {
		log.Panic(err.Error())
		return
	}

	for idx := range s.infos {
		if err = writer.Write(_type.SpanInfoTableRow2Str(&s.infos[idx])); err != nil {
			log.Panic(err.Error())
			return
		}
	}
	writer.Flush()
}

func (s *SpanVis) PrepareData(tt OpType) {
	if _type.SourceFile == "" {
		s.db.Table("span_info").Where(fmt.Sprintf("span_kind='%s'", tt.String())).Find(&s.infos)
		// here, we also save the data read from db as CSV file
		s.saveCSV(tt)
	} else {
		// decode data from file
		s.decodeCSV(tt)
	}

	condition := ""
	if tt == S3FSOperation {
		condition = "S3FS"
	} else {
		condition = "LocalFS"
	}

	for idx, _ := range s.infos {
		var extra map[string]interface{}
		if err := json.Unmarshal([]byte(s.infos[idx].Extra), &extra); err != nil {
			fmt.Println(fmt.Errorf(err.Error()))
		}

		if strings.HasSuffix(extra["name"].(string), ".csv") {
			continue
		}

		if s.infos[idx].SpanName == fmt.Sprintf("%s.Write", condition) {
			s.writeRecords = append(s.writeRecords, s.infos[idx])
		} else if s.infos[idx].SpanName == fmt.Sprintf("%s.read", condition) {
			s.readRecords = append(s.readRecords, s.infos[idx])
		}
	}
}

func (s *SpanVis) visualize(tt OpType) {
	s.PrepareData(tt)
	s.visualize_ObjReqHeatmap(tt)
	s.visualize_ObjReqThroughTime(tt, time.Second*5)
	s.visualize_ObjReadLatency(tt)
}

func (s *SpanVis) parseTNAndCN(data []_type.SpanInfoTable) (cnInfo, tnInfo []_type.SpanInfoTable) {
	for idx := range data {
		if data[idx].NodeType == "CN" {
			cnInfo = append(cnInfo, data[idx])
		} else {
			// TN
			tnInfo = append(tnInfo, data[idx])
		}
	}
	return
}

// visualize_ObjReadLatency records the spent time on every obj requests in time order.
// the X-axis:	a prefix of obj name + request end time
// the Y-axis:	time spend in millisecond
func (s *SpanVis) visualize_ObjReadLatency(tt OpType) {
	title := "Read"
	data := s.readRecords

	sort.Slice(data, func(i, j int) bool {
		return data[i].EndTime.Before(data[j].EndTime)
	})

	cnInfo, tnInfo := s.parseTNAndCN(data)
	getData := func(info []_type.SpanInfoTable) (values []float64, labels []string) {
		for idx := range info {
			extra := s.unmarshExtra(info[idx].Extra)
			name := extra["name"].(string)
			labels = append(labels, strings.Split(name, "-")[0]+" # "+info[idx].EndTime.String())
			// time.Millisecond
			values = append(values, float64(info[idx].Duration/(1000*1000)))
		}
		return
	}

	appendData := func(st string, labels []string, values []float64) {
		spanObjReadLatencyData = append(spanObjReadLatencyData, html.SignalLinePageData{
			Labels: labels,
			Values: values,
			XAxis:  "时间戳",
			YAxis:  "时延 (ms)",
			Title:  st + "  " + tt.String() + "  " + title + ":  Obj Read Latency",
		})
	}

	values, labels := getData(cnInfo)
	appendData("CN", labels, values)

	values, labels = getData(tnInfo)
	appendData("TN", labels, values)

	s.appendToRenderData(spanObjReadLatencyData)

}

func (s *SpanVis) visualize_ObjReqThroughTime(tt OpType, duration time.Duration) {
	// show object request num in every duration
	title := []string{"Read", "Write"}
	data := [][]_type.SpanInfoTable{s.readRecords, s.writeRecords}
	for round := range data {
		sort.Slice(data[round], func(i, j int) bool {
			return data[round][i].EndTime.Before(data[round][j].EndTime)
		})

		cnInfo, tnInfo := s.parseTNAndCN(data[round])
		getData := func(info []_type.SpanInfoTable) ([]float64, []string) {
			var cntByDuration []float64
			var endTime []string

			if len(info) == 0 {
				return cntByDuration, endTime
			}

			last := info[0].EndTime
			idx := 0
			for idx < len(info) {
				cnt := int64(0)
				for idx < len(info) && info[idx].EndTime.Sub(last) <= duration {
					cnt++
					idx++
				}
				if idx < len(info) {
					last = info[idx].EndTime
				}
				if cnt == 0 {
					continue
				}

				endTime = append(endTime, info[idx-1].EndTime.String())
				cntByDuration = append(cntByDuration, float64(cnt))
			}
			return cntByDuration, endTime
		}

		appendData := func(st string, labels []string, values []float64) {
			spanObjReqThroughTimeData = append(spanObjReqThroughTimeData, html.SignalLinePageData{
				Labels: labels,
				Values: values,
				XAxis:  "时间戳",
				YAxis:  "object 访问数量",
				Title:  st + "  " + tt.String() + "  " + title[round] + ":  Obj Req Through Time",
			})
		}
		values, labels := getData(cnInfo)
		appendData("CN", labels, values)

		values, labels = getData(tnInfo)
		appendData("TN", labels, values)

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
	title := []string{"Read", "Write"}
	data := [][]_type.SpanInfoTable{s.readRecords, s.writeRecords}

	for round, _ := range data {
		cnInfo, tnInfo := s.parseTNAndCN(data[round])

		getData := func(info []_type.SpanInfoTable) (values []float64, labels []string) {
			var objName []string
			name2Cnt := make(map[string]float64)

			for idx, _ := range info {
				extra := s.unmarshExtra(info[idx].Extra)
				name := extra["name"].(string)
				_, ok := name2Cnt[name]
				name2Cnt[name]++
				if !ok {
					objName = append(objName, extra["name"].(string))
				}
			}
			sort.Slice(objName, func(i, j int) bool {
				return name2Cnt[objName[i]] > name2Cnt[objName[j]]
			})

			for idx, _ := range objName {
				labels = append(labels, objName[idx])
				values = append(values, name2Cnt[objName[idx]])
			}

			return
		}

		appendData := func(st string, labels []string, values []float64) {
			spanObjReqHeatmapData = append(spanObjReqHeatmapData, html.SignalLinePageData{
				Labels: labels,
				Values: values,
				XAxis:  "object name",
				YAxis:  "object 访问数量",
				Title:  st + "  " + tt.String() + "  " + title[round] + ":  Obj Request Heatmap",
			})
		}

		values, labels := getData(cnInfo)
		appendData("CN", labels, values)

		values, labels = getData(tnInfo)
		appendData("TN", labels, values)
	}
	s.appendToRenderData(spanObjReqHeatmapData)
}

func (s *SpanVis) appendToRenderData(data []html.SignalLinePageData) {
	renderData.Data = append(renderData.Data, data...)
}
