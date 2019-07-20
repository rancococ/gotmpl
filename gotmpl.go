package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
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
	build       = ""
	version     = "1.0.0"
	author      = "rancococ@qq.com"
	argHelp     = false
	argVerbose  = false
	argVersion  = false
	argTemplate string
	argJsondata string
	argYamldata string
	argOutfile  string
)

const (
	TypeString = "s"
	TypeFile   = "f"
)

func init() {
	//log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	getopt.FlagLong(&argHelp, "help", 'h', "Show help info.")
	getopt.FlagLong(&argVerbose, "verbose", 0, "Output debug info.")
	getopt.FlagLong(&argVersion, "version", 'v', "Output version info.")
	getopt.FlagLong(&argTemplate, "template", 't', "Template info, support string and file, start with [s:] or [f:].\nFormat:s:string|f:filepath", "s:string|f:filepath")
	getopt.FlagLong(&argJsondata, "jsondata", 'j', "Jsondata info, support string and file, start with [s:] or [f:].\nFormat:s:string|f:filepath", "s:string|f:filepath")
	getopt.FlagLong(&argYamldata, "yamldata", 'y', "Yamldata info, support string and file, start with [s:] or [f:].\nFormat:s:string|f:filepath", "s:string|f:filepath")
	getopt.FlagLong(&argOutfile, "outfile", 'o', "Out file path.\nformat:/path/out.txt", "/path/out.txt")
	getopt.ParseV2()
	/**
	flag.BoolVar(&argHelp, "help", false, "Show help info.")
	flag.BoolVar(&argVerbose, "verbose", false, "Output debug info.")
	flag.BoolVar(&argVersion, "version", false, "Output version info.")
	flag.StringVar(&argTemplate, "template", "", "Template info, support string and file, start with [s:] or [f:].\nFormat:`s:string|f:filepath`")
	flag.StringVar(&argJsondata, "jsondata", "", "Jsondata info, support string and file, start with [s:] or [f:].\nFormat:`s:string|f:filepath`")
	flag.StringVar(&argYamldata, "yamldata", "", "Yamldata info, support string and file, start with [s:] or [f:].\nFormat:`s:string|f:filepath`")
	flag.StringVar(&argOutfile, "outfile", "", "Out file path.\nformat:`/abc/xyz.txt`")
	flag.Parse()
	*/
}

/**
usage of this command tool:
    -h, --help
    -t, --template="s:xxx" or --template="f:/path/xxx.tpl"
    -j, --jsondata="s:xxx" or --jsondata="f:/path/xxx.json"
    -y, --yamldata="s:xxx" or --yamldata="f:/path/xxx.yaml"
    -o, --outfile="/path/xxx.txt"
    --verbose
    -v, --version
*/
func main() {

	defer func() {
		if r := recover(); r != nil {
			logPrintfError("error : %s", r)
			time.Sleep(100 * time.Millisecond)
			os.Exit(1)
		}
	}()

	if len(os.Args) == 1 || argHelp {
		//flag.Usage()
		getopt.Usage()
		os.Exit(0)
	}

	if argVersion {
		showVersion()
		os.Exit(0)
	}

	if argTemplate == "" {
		fmt.Println("[--template] parameters must be entered.")
		os.Exit(1)
	}

	if (argJsondata == "" && argYamldata == "") || (argJsondata != "" && argYamldata != "") {
		fmt.Println("[--jsondata] and [--yamldata] must enter one of them.")
		os.Exit(1)
	}

	var err error
	var dataobject map[string]interface{}
	var outwriter io.Writer

	// 解析template参数
	tmplcontent, err := parseTemplate(argTemplate)
	if err != nil {
		panic(err.Error())
	}
	// 解析jsondata参数
	if argJsondata != "" {
		dataobject, err = parseJsondata(argJsondata)
		if err != nil {
			panic(err.Error())
		}
	}
	// 解析yamldata参数
	if argYamldata != "" {
		dataobject, err = parseYamldata(argYamldata)
		if err != nil {
			panic(err.Error())
		}
	}

	// 解析out参数
	if argOutfile != "" {
		logPrintf("\n")
		logPrintf("%20v : %v", "out file path", argOutfile)
		// create out file
		outepath := path.Dir(argOutfile)
		err = os.MkdirAll(outepath, 0777)
		if err != nil {
			panic(err.Error())
		}
		outfile, err := os.OpenFile(argOutfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
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
		outwriter = io.MultiWriter(writers...)
	} else {
		writers := []io.Writer{
			os.Stdout,
		}
		outwriter = io.MultiWriter(writers...)
	}
	logPrintf("%20v : %v", "out file content", "\n")

	// 渲染模板
	tmplobject := template.Must(template.New("tmpl").Parse(tmplcontent))
	err = tmplobject.Execute(outwriter, dataobject)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println()
	logPrintf("success!")

}

// 查看版本信息
func showVersion() {
	fmt.Printf("version : %v\n", version)
	fmt.Printf("build   : %v\n", build)
	fmt.Printf("author  : %v\n", author)
}

// 解析template参数
func parseTemplate(argTemplate string) (content string, err error) {
	logPrintf("\n")
	logPrintf("%20v : %v", "parse parameter", "template")
	tmpltype := SubStringBefore(argTemplate, 1)
	switch tmpltype {
	case TypeString:
		content = SubStringAfter(argTemplate, 2)
		logPrintf("%20v : %v", "template type", "string")
		logPrintf("%20v : %v", "template content", "\n"+content)
	case TypeFile:
		filepath := SubStringAfter(argTemplate, 2)
		logPrintf("%20v : %v", "template type", "file")
		logPrintf("%20v : %v", "template filepath", filepath)
		filebytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			return "", err
		}
		content = string(filebytes)
		logPrintf("%20v : %v", "template content", "\n"+content)
	}
	logPrintf("***************************************************************************")
	return content, nil
}

// 解析jsondata参数
func parseJsondata(argJsondata string) (obj map[string]interface{}, err error) {
	logPrintf("\n")
	logPrintf("%20v : %v", "parse parameter", "jsondata")
	var content string
	datatype := SubStringBefore(argJsondata, 1)
	switch datatype {
	case TypeString:
		content = SubStringAfter(argJsondata, 2)
		logPrintf("%20v : %v", "jsondata type", "string")
		logPrintf("%20v : %v", "jsondata content", "\n"+content)
	case TypeFile:
		filepath := SubStringAfter(argJsondata, 2)
		logPrintf("%20v : %v", "jsondata type", "file")
		logPrintf("%20v : %v", "jsondata filepath", filepath)
		filebytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		content = string(filebytes)
		logPrintf("%20v : %v", "jsondata content", "\n"+content)
	}
	// json转map
	obj, err = JsonToMap(content)
	if err != nil {
		return nil, err
	}
	logPrintf("%20v : %v", "object content", obj)
	logPrintf("***************************************************************************")
	return obj, nil
}

// 解析yamldata参数
func parseYamldata(argYamldata string) (obj map[string]interface{}, err error) {
	logPrintf("\n")
	logPrintf("%20v : %v", "parse parameter", "yamldata")
	var content string
	datatype := SubStringBefore(argYamldata, 1)
	switch datatype {
	case TypeString:
		content = SubStringAfter(argYamldata, 2)
		logPrintf("%20v : %v", "yamldata type", "string")
		logPrintf("%20v : %v", "yamldata content", content)
	case TypeFile:
		filepath := SubStringAfter(argYamldata, 2)
		logPrintf("%20v : %v", "yamldata type", "file")
		logPrintf("%20v : %v", "yamldata filepath", filepath)
		filebytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		content = string(filebytes)
		logPrintf("%20v : %v", "yamldata content", content)
	}
	// yaml转map
	obj, err = YamlToMap(content)
	if err != nil {
		return nil, err
	}
	logPrintf("%20v : %v", "object content", obj)
	logPrintf("***************************************************************************")
	return obj, nil
}

func logPrintf(format string, v ...interface{}) {
	if !argVerbose {
		return
	}
	if len(v) > 0 {
		log.Printf(format, v...)
	} else {
		log.Printf(format)
	}
}

func logPrintfError(format string, v ...interface{}) {
	if len(v) > 0 {
		log.Printf(format, v...)
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

// Json字符串转Map对象
func JsonToMap(jsonstr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonstr), &data)
	return data, err
}

// Yaml字符串转Map对象
func YamlToMap(yamlstr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlstr), &data)
	return data, err
}
