package _type

import "time"

type LogInfoTable struct {
	TraceId    string    `gorm:"column:trace_id"`
	SpanId     string    `gorm:"column:span_id"`
	SpanKind   string    `gorm:"column:span_kind"`
	NodeUuid   string    `gorm:"column:node_uuid"`
	NodeType   string    `gorm:"column:node_type"`
	Timestamp  time.Time `gorm:"column:timestamp"`
	LoggerName string    `gorm:"column:logger_name"`
	Level      string    `gorm:"column:level"`
	Caller     string    `gorm:"column:caller"`
	Message    string    `gorm:"column:message"`
	Extra      string    `gorm:"column:extra"`
	Stack      string    `gorm:"column:stack"`
}
