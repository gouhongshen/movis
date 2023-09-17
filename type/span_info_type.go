package _type

import (
	"fmt"
	"strconv"
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

func (s *SpanInfoTable) SetVal(name string, val string) {
	switch name {
	case "trace_id":
		s.TraceId = val
	case "span_id":
		s.SpanId = val
	case "parent_span_id":
		s.ParentSpanId = val
	case "span_kind":
		s.SpanKind = val
	case "node_uuid":
		s.NodeUUID = val
	case "node_type":
		s.NodeType = val
	case "span_name":
		s.SpanName = val
	case "start_time":
		var err error
		s.StartTime, err = time.Parse("2006-01-02 15:04:05.999999", val)
		if err != nil {
			fmt.Println("parse start time: ", err.Error())
		}
	case "end_time":
		var err error
		s.EndTime, err = time.Parse("2006-01-02 15:04:05.999999", val)
		if err != nil {
			fmt.Println("parse end time: ", err.Error())
		}
	case "duration":
		dur, err := strconv.Atoi(val)
		if err != nil {
			fmt.Println(err.Error())
		}
		s.Duration = int64(dur)
	case "resource":
		s.Resource = val
	case "extra":
		s.Extra = val
	default:
		panic("no such name")
	}
}

func SpanInfoTableRow2Str(info *SpanInfoTable) []string {
	start := info.StartTime.Format("2006-01-02 15:04:05.999999")
	end := info.EndTime.Format("2006-01-02 15:04:05.999999")

	return []string{
		info.TraceId, info.SpanId, info.ParentSpanId, info.SpanKind,
		info.NodeUUID, info.NodeType, info.SpanName, start,
		end, strconv.FormatInt(info.Duration, 10),
		info.Resource, info.Extra,
	}

}
