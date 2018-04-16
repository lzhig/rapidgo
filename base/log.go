/**
 * @author [Bruce]
 * @email [lzhig@outlook.com]
 * @create date 2018-03-15 09:32:41
 * @modify date 2018-03-15 09:32:41
 * @desc [description]
 */

package base

import (
	"bytes"
	"flag"
	"os"
	"runtime"

	"github.com/golang/glog"
)

// LogInit 指定日志目录
func LogInit(logDir string) {
	// 设置log目录
	flag.CommandLine.Set("log_dir", logDir)
	os.Mkdir(logDir, os.ModePerm)
}

// LogInfo function
func LogInfo(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

// LogWarn function
func LogWarn(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

// LogError function
func LogError(args ...interface{}) {
	glog.ErrorDepth(1, args...)
}

// LogFatal function
func LogFatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

// LogFlush function
func LogFlush() {
	glog.Flush()
}

// LogPanic function
func LogPanic() {
	if err := recover(); err != nil {
		LogError("panic:", err) // 这里的err其实就是panic传入的内容，55
		LogError(string(panicTrace(10)))
	}
}

// PanicTrace trace panic stack info.
func panicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}
