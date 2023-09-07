package script

import (
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
	"sort"
	"strings"
	"time"
)

type OpType int

const (
	S3FSOperation    OpType = 0
	LocalFSOperation OpType = 1
)

var opType2Name = map[OpType]string{
	S3FSOperation:    "s3FSOperation",
	LocalFSOperation: "localFSOperation",
}

func (op *OpType) String() string {
	return opType2Name[*op]
}

var renderData html.LinePageData

type SpanVis struct {
	db                        *gorm.DB
	infos                     []_type.SpanInfoTable
	readRecords, writeRecords []_type.SpanInfoTable
}

var spanVis *SpanVis

func spanVisInit() {
	defer func() {
		spanVis.infos = make([]_type.SpanInfoTable, 0)
		spanVis.readRecords = make([]_type.SpanInfoTable, 0)
		spanVis.writeRecords = make([]_type.SpanInfoTable, 0)
	}()

	if spanVis != nil {
		return
	}

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

func (s *SpanVis) tmplInject(w http.ResponseWriter, req *http.Request) {
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

func S3FSOperationHandler(w http.ResponseWriter, req *http.Request) {
	spanVisInit()
	spanVis.visualize(S3FSOperation)
	spanVis.tmplInject(w, req)
}

func (s *SpanVis) PrepareData(tt OpType) {
	if _type.SourceFile == "" {
		s.db.Table("span_info").Where(fmt.Sprintf("span_kind='%s'", tt.String())).Find(&s.infos)
	} else {
		// decode data from file
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
	s.visualize_ObjReqThroughTime(tt)
}

func (s *SpanVis) visualize_ObjReqThroughTime(tt OpType) {
	// show object request num in every second
	title := []string{"Read", "Write"}
	data := [][]_type.SpanInfoTable{s.readRecords, s.writeRecords}
	for round := range data {
		sort.Slice(data[round], func(i, j int) bool {
			return data[round][i].EndTime.Before(data[round][j].EndTime)
		})

		if len(data[round]) == 0 {
			continue
		}

		var cntByDuration []float64
		var endTime []string
		last := data[round][0].EndTime

		idx := 0
		for idx < len(data[round]) {
			cnt := int64(0)
			for idx < len(data[round]) && data[round][idx].EndTime.Sub(last) <= time.Second {
				cnt++
				idx++
			}

			if idx < len(data[round]) {
				last = data[round][idx].EndTime
			}

			endTime = append(endTime, data[round][idx-1].EndTime.String())
			cntByDuration = append(cntByDuration, float64(cnt))
		}

		var labels []string
		var values []float64

		for idx, _ = range endTime {
			labels = append(labels, endTime[idx])
			values = append(values, cntByDuration[idx])
		}

		s.appendToRenderData(labels, values,
			"时间", "访问数量", tt.String()+" "+title[round]+": ObjReqThroughTime")

	}
}

func (s *SpanVis) visualize_ObjReqHeatmap(tt OpType) {
	title := []string{"Read", "Write"}
	data := [][]_type.SpanInfoTable{s.readRecords, s.writeRecords}

	for round, _ := range data {
		var objName []string
		name2Cnt := make(map[string]float64)

		for idx, _ := range data[round] {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(data[round][idx].Extra), &extra); err != nil {
				fmt.Println(fmt.Errorf(err.Error()))
			}

			name := extra["name"].(string)
			//if strings.HasSuffix(name, ".csv") {
			//	continue
			//}

			_, ok := name2Cnt[name]
			name2Cnt[name]++

			if !ok {
				objName = append(objName, extra["name"].(string))
			}
		}

		sort.Slice(objName, func(i, j int) bool {
			return name2Cnt[objName[i]] > name2Cnt[objName[j]]
		})

		var labels []string
		var values []float64

		for idx, _ := range objName {
			labels = append(labels, objName[idx])
			values = append(values, name2Cnt[objName[idx]])
		}

		s.appendToRenderData(labels, values,
			"object name", "访问数量", tt.String()+" "+title[round]+": ObjRequestHeatmap")
	}
}

func (s *SpanVis) appendToRenderData(
	labels []string, values []float64,
	xaxis string, yaxis string, title string) {

	renderData.Data = append(renderData.Data,
		struct {
			Labels []string
			Values []float64
			XAxis  string
			YAxis  string
			Title  string
		}{
			Labels: labels,
			Values: values,
			XAxis:  xaxis,
			YAxis:  yaxis,
			Title:  title,
		},
	)
}
