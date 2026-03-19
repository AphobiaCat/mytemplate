package log

import (
	"fmt"
	"path/filepath"
	"runtime"

	"mytemplate/internal/global"
)

func DebugLog(v ...interface{}) {

	if global.AppConfig.Env == "prod" {
		return
	}

	Log(v...)
}

func DebugError(v ...interface{}) {

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Failed to get caller information")
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		fmt.Println("Failed to get function information")
		return
	}

	path := file
	filename := filepath.Base(path)

	var outputStr string = "[error] file[" + filename + "]\t| func[" + fn.Name() + "]\t| line[" + fmt.Sprintf("%v", line) + "]\t| log:"

	for _, val := range v {
		outputStr += fmt.Sprintf("%v", val)
	}

	fmt.Println("\033[1;31m" + outputStr + "\033[0m")
}

func Log(v ...interface{}) {

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Failed to get caller information")
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		fmt.Println("Failed to get function information")
		return
	}

	path := file
	filename := filepath.Base(path)

	var outputStr string = "[info] file[" + filename + "]\t| func[" + fn.Name() + "]\t| line[" + fmt.Sprintf("%v", line) + "]\t| log:"

	for _, val := range v {
		outputStr += fmt.Sprintf("%v", val)
	}

	fmt.Println(outputStr)

}
