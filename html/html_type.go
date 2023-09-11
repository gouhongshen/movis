package html

type SignalLinePageData struct {
	Labels    []string
	Values    []float64
	XAxis     string
	YAxis     string
	Title     string
	ChartType string
}

type LinePageData struct {
	Data []SignalLinePageData
}
