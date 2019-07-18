package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
	"time"
)

var (
	build        = ""
	version      = "1.0.0"
	author       = "rancococ@qq.com"
	tmpl         = flag.String("tmpl", "", "s:template string/f:template file path")
	data         = flag.String("data", "", "s:yaml string/f:yaml file path")
	out          = flag.String("out", "", "out file path")
	quiet        = flag.Bool("quiet", false, "quiet")
	help         = flag.Bool("help", false, "show help")
	h            = flag.Bool("h", false, "show help")
	tmpltype     string
	datatype     string
	tmplcontent  string
	tmplfilepath string
	datacontent  string
	datafilepath string
)

const (
	TypeString = "s"
	TypeFile   = "f"
)

func init() {
	//log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	flag.Parse()
}

// args: --tmpl="s:xxx" --tmpl="f:xxx.tpl" --data="s:xxx"  --data="f:xxx.txt"
func main() {

	defer func() {
		if r := recover(); r != nil {
			logPrintfError("error : %s", r)
			time.Sleep(1 * time.Second)
		}
	}()

	if len(os.Args) == 1 || *h || *help {
		flag.Usage()
		os.Exit(0)
	}

	// show build info
	logPrintf("***************************************************************************")
	logPrintf("version : %v", version)
	logPrintf("build   : %v", build)
	logPrintf("author  : %v", author)
	logPrintf("args    : %v", os.Args)
	logPrintf("***************************************************************************")

	// 解析tmpl参数
	logPrintf("\n")
	logPrintf("parse parameter : tmpl")
	tmpltype = SubStringBefore(*tmpl, 1)
	switch tmpltype {
	case TypeString:
		tmplcontent = SubStringAfter(*tmpl, 2)
		logPrintf("tmpl type    : %s", "string")
		logPrintf("tmpl content : %s", tmplcontent)
	case TypeFile:
		tmplfilepath = SubStringAfter(*tmpl, 2)
		logPrintf("tmpl type     : %s", "file")
		logPrintf("tmpl filepath : %s", tmplfilepath)
		content, err := ioutil.ReadFile(tmplfilepath)
		if err != nil {
			panic(err.Error())
		}
		tmplcontent = string(content)
		logPrintf("tmpl content : %s", tmplcontent)
	}
	// 解析data参数
	logPrintf("\n")
	logPrintf("parse parameter : data")
	datatype = SubStringBefore(*data, 1)
	switch datatype {
	case TypeString:
		datacontent = SubStringAfter(*data, 2)
		logPrintf("data type    : %s", "string")
		logPrintf("data content : %s", datacontent)
	case TypeFile:
		datafilepath = SubStringAfter(*data, 2)
		logPrintf("data type     : %s", "file")
		logPrintf("data filepath : %s", datafilepath)
		content, err := ioutil.ReadFile(datafilepath)
		if err != nil {
			panic(err.Error())
		}
		datacontent = string(content)
		logPrintf("data content : %s", datacontent)
	}
	// 解析out参数
	logPrintf("\n")
	logPrintf("parse parameter : out")
	logPrintf("out filepath : %s", *out)

	// yaml转map
	logPrintf("\n")
	dataobject, err := YamlToMap(datacontent)
	if err != nil {
		panic(err.Error())
	}
	logPrintf("data content : %s", dataobject)

	var multiWriter io.Writer
	if *out != "" {
		// create out file
		logPrintf("\n")
		outepath := path.Dir(*out)
		err = os.MkdirAll(outepath, 0777)
		if err != nil {
			panic(err.Error())
		}
		outfile, err := os.OpenFile(*out, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
		if err != nil {
			panic(err.Error())
		}
		defer func() {
			_ = outfile.Close()
		}()
		writers := []io.Writer{
			outfile,
			os.Stdout,
		}
		multiWriter = io.MultiWriter(writers...)
	} else {
		writers := []io.Writer{
			os.Stdout,
		}
		multiWriter = io.MultiWriter(writers...)
	}
	logPrintf("out content : \n")
	// 渲染模板
	tmplobject := template.Must(template.New("tmpl").Parse(tmplcontent))
	err = tmplobject.Execute(multiWriter, dataobject)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println()
	logPrintf("success!")

}

func logPrintf(format string, v ...interface{}) {
	if *quiet {
		return
	}
	if len(v) > 0 {
		log.Printf(format, v)
	} else {
		log.Printf(format)
	}
}

func logPrintfError(format string, v ...interface{}) {
	if len(v) > 0 {
		log.Printf(format, v)
	} else {
		log.Printf(format)
	}
}

// 截取字符串 end 之前(不包括)(下标从0开始)
func SubStringBefore(str string, end int) string {
	rs := []rune(str)
	length := len(rs)
	if end < 0 {
		end = 0
	}
	if end >= length {
		end = length
	}
	return string(rs[0:end])
}

// 截取字符串 start 之后(包括)(下标从0开始)
func SubStringAfter(str string, start int) string {
	rs := []rune(str)
	length := len(rs)
	if start < 0 {
		start = 0
	}
	if start >= length {
		start = length
	}
	return string(rs[start:length])
}

// Yaml字符串转Map对象
func YamlToMap(yamlstr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlstr), &data)
	return data, err
}