package main

import (
	"fmt"
	"movis/script"
	_type "movis/type"
	"net/http"
	"os"
	"strings"
)

func main() {
	decodeArgs(os.Args)
	fillDefault()

	defer release()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/span_info", spanRoot)
	http.HandleFunc("/span_info/s3_fs_operation", script.S3FSOperationHandler)
	//http.HandleFunc("/log_info", script.VisLogInfoHandler)

	fmt.Printf("Server started at :%s\n", _type.DstPort)
	if err := http.ListenAndServe(":"+_type.DstPort, nil); err != nil {
		fmt.Println(err.Error())
	}
}

func release() {

}

func fillDefault() {
	if _type.DstPort == "" {
		_type.DstPort = "11235"
	}

	if _type.SrcPort == "" {
		_type.SrcPort = "6001"
	}

	if _type.SrcHost == "" {
		_type.SrcHost = "127.0.0.1"
	}

	if _type.SrcPassword == "" {
		_type.SrcPassword = "111"
	}

	if _type.SrcUsrName == "" {
		_type.SrcUsrName = "dump"
	}

}

// type 1: -http=:dstPort -hSrcHost -PSrcPort -uSrcUsrName -pSrcPwd
// type 2: -f srcFile
const (
	ArgsFormat1 = 3
	ArgsFormat2 = 6
)

func decodeArgs(args []string) bool {
	if len(args) == ArgsFormat1 {
		if args[1] != "-f" {
			return false
		}
		_type.SourceFile = args[2]
		return true
	} else if len(args) == ArgsFormat2 {
		idx := map[string]*string{
			"-http=:": &_type.DstPort,
			"-h":      &_type.SrcHost,
			"-p":      &_type.SrcPassword,
			"-u":      &_type.SrcUsrName,
			"-P":      &_type.SrcPort,
		}

		for p, o := range idx {
			curArg := ""
			for _, arg := range args {
				if strings.HasPrefix(arg, p) {
					curArg = arg
				}
			}
			if curArg == "" {
				return false
			}
			*o = strings.Trim(curArg, p)
		}

		return true
	}

	return false
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>MO VIS</title>
    </head>
    <body>
        <ul>
            <li><a href="/span_info"> Span Info </a></li>
            <li><a href="/log_info"> Log Info </a></li>
        </ul>
    </body>
    </html>
    `
	_, err := w.Write([]byte(html))
	if err != nil {
		panic(err.Error())
	}
}

func spanRoot(w http.ResponseWriter, req *http.Request) {
	html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>SPAN VIS</title>
    </head>
    <body>
        <ul>
            <li><a href="/span_info/local_fs_operation"> Local FS Operation </a></li>
            <li><a href="/span_info/s3_fs_operation"> S3 FS Operation </a></li>
        </ul>
    </body>
    </html>
    `
	_, err := w.Write([]byte(html))
	if err != nil {
		panic(err.Error())
	}
}
