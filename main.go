package main

import (
	"fmt"
	"log"
	"movis/script"
	_type "movis/type"
	"net/http"
	"os"
)

func main() {
	decodeArgs()
	_type.FillTypeDefault()

	defer release()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/span_info", spanRoot)
	http.HandleFunc("/span_info/s3_fs_operation", script.S3FSOperationHandler)
	http.HandleFunc("/span_info/local_fs_operation", script.LocalFSOperationHandler)

	fmt.Printf("Server started at :%s\n", _type.DstPort)
	if err := http.ListenAndServe(":"+_type.DstPort, nil); err != nil {
		fmt.Println(err.Error())
	}
}

func release() {

}

// host, port, usr name, pwd all have the default value, for the details see type.go
// but if user specified the -f argument，it will only read data from file which follows that parameter.
// the allowed format like this:
//
//	-x y -r s
//
// this format is not allowed:
//
//	-xy -rs
func decodeArgs() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}

	name2Args := map[string]*string{
		"-http": &_type.DstPort,
		"-h":    &_type.SrcHost,
		"-p":    &_type.SrcPassword,
		"-u":    &_type.SrcUsrName,
		"-P":    &_type.SrcPort,
		"-f":    &_type.SourceFile,
	}
	// -h, -p, -u, -P are mutually exclusive with -f

	if len(args)%2 != 0 {
		panic("expecting an even number of parameters")
	}

	x, y := false, false
	for i := 0; i < len(args); {
		if _, ok := name2Args[args[i]]; !ok {
			log.Panicf("no such parameter: %s", args[i])
		}

		if args[i] == "-http" {
			*name2Args["-http"] = args[i+1]

		} else if args[i] == "-f" {
			if y == true {
				panic("-h, -p, -u, -P are mutually exclusive with -f")
			}
			x = true
			*name2Args["-f"] = args[i+1]

		} else {
			if x == true {
				panic("-h, -p, -u, -P are mutually exclusive with -f")
			}
			y = true
			*name2Args[args[i]] = args[i+1]
		}
		i += 2
	}
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
