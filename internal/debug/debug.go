package debug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const flags = log.Ltime | log.Lmicroseconds

type Logger struct {
	l       *log.Logger
	enabled bool
	prefix  string
}

func NewLogger(prefix string) *Logger {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}
	return &Logger{l: log.New(os.Stderr, "", flags), prefix: prefix, enabled: true}
}

func NewLoggerIf(prefix string, enabled bool) *Logger {
	logger := NewLogger(prefix)
	logger.enabled = enabled
	return logger
}

func (dbg *Logger) Output(s string) {
	if !Enabled || !dbg.enabled {
		return
	}

	var callerInfo string
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		var funcName string
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
			i := strings.LastIndexByte(funcName, '.')
			if i >= 0 {
				funcName = funcName[i+1:]
			}
		} else {
			funcName = "???"
		}
		callerInfo = fmt.Sprintf("%s:%d(%s): ", filepath.Base(file), line, funcName)
	} else {
		callerInfo = "???: "
	}
	dbg.l.Output(0, fmt.Sprintf("%s%s", callerInfo, s))
}

func (dbg *Logger) Printf(format string, v ...any) {
	if Enabled && dbg.enabled {
		dbg.Output(fmt.Sprintf(format, v...))
	}
}

func (dbg *Logger) Println(v ...any) {
	if Enabled && dbg.enabled {
		dbg.Output(fmt.Sprint(v...))
	}
}
