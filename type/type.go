package _type

var SourceFile string
var DstPort string
var SrcPort string
var SrcHost string
var SrcUsrName string
var SrcPassword string
var ReportFile string

const ReportsRootDir string = "./data/reports"
const SrcDataDir string = "./data/records"

func FillTypeDefault() {
	//if DstPort == "" {
	//	DstPort = "11235"
	//}

	if SrcPort == "" {
		SrcPort = "6001"
	}

	if SrcHost == "" {
		SrcHost = "127.0.0.1"
	}

	if SrcPassword == "" {
		SrcPassword = "111"
	}

	if SrcUsrName == "" {
		SrcUsrName = "dump"
	}

}
