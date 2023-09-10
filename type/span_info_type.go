package _type

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const SpanReportDir = ReportsRootDir + "/" + "span_info/"

type SpanInfoTable struct {
	TraceId      string    `gorm:"column:trace_id"`
	SpanId       string    `gorm:"column:span_id"`
	ParentSpanId string    `gorm:"column:parent_span_id"`
	SpanKind     string    `gorm:"column:span_kind"`
	NodeUUID     string    `gorm:"column:node_uuid"`
	NodeType     string    `gorm:"column:node_type"`
	SpanName     string    `gorm:"column:span_name"`
	StartTime    time.Time `gorm:"column:start_time"`
	EndTime      time.Time `gorm:"column:end_time"`
	// nanoseconds
	Duration int64  `gorm:"column:duration"`
	Resource string `gorm:"column:resource"`
	Extra    string `gorm:"column:extra"`
}

func SpanInfoTableCSVHead() []string {
	return []string{
		"trace_id", "span_id", "parent_span_id", "span_kind",
		"node_uuid", "node_type", "span_name", "start_time", "end_time",
		"duration", "resource", "extra",
	}
}

func SpanInfoTableRow2Str(info *SpanInfoTable) []string {
	return []string{
		info.TraceId, info.SpanId, info.ParentSpanId, info.SpanKind,
		info.NodeUUID, info.NodeType, info.SpanName, info.StartTime.String(),
		info.EndTime.String(), strconv.FormatInt(info.Duration, 10),
		info.Resource, info.Extra,
	}
}

func SpanInfoMemberSetFunc() map[string]func(string) reflect.Value {
	funcs := map[string]func(string) reflect.Value{
		"trace_id":       func(src string) reflect.Value { return reflect.ValueOf(src) },
		"span_id":        func(src string) reflect.Value { return reflect.ValueOf(src) },
		"parent_span_id": func(src string) reflect.Value { return reflect.ValueOf(src) },
		"span_kind":      func(src string) reflect.Value { return reflect.ValueOf(src) },
		"node_uuid":      func(src string) reflect.Value { return reflect.ValueOf(src) },
		"node_type":      func(src string) reflect.Value { return reflect.ValueOf(src) },
		"span_name":      func(src string) reflect.Value { return reflect.ValueOf(src) },
		"start_time": func(src string) reflect.Value {
			datetime, _ := time.Parse("2006-01-02 15:04:05", src)
			return reflect.ValueOf(datetime)
		},
		"end_time": func(src string) reflect.Value {
			datetime, _ := time.Parse("2006-01-02 15:04:05", src)
			return reflect.ValueOf(datetime)
		},
		"duration": func(src string) reflect.Value {
			val, err := strconv.Atoi(src)
			if err != nil {
				fmt.Println(err.Error())
			}
			return reflect.ValueOf(int64(val))
		},
		"resource": func(src string) reflect.Value { return reflect.ValueOf(src) },
		"extra":    func(src string) reflect.Value { return reflect.ValueOf(src) },
	}

	return funcs
}

func SpanInfoMapTag2Values(s *SpanInfoTable) map[string]reflect.Value {
	ret := make(map[string]reflect.Value)
	values := reflect.ValueOf(s).Elem()
	fields := reflect.TypeOf(*s)
	for i := 0; i < values.NumField(); i++ {
		v := values.Field(i)
		tag := fields.Field(i).Tag.Get("gorm")
		ret[strings.TrimPrefix(tag, "column:")] = v
	}
	return ret
}
