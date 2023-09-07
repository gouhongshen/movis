package html

type LinePageData struct {
	//Head string
	Data []struct {
		Labels []string
		Values []float64
		XAxis  string
		YAxis  string
		Title  string
	}
}
