package debug

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Default debug logger flags.
const flags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix

type Logger struct {
	l *log.Logger
}

func NewLogger(prefix string) *Logger {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}
	return &Logger{l: log.New(os.Stderr, prefix, flags)}
}

func (dbg *Logger) Printf(format string, v ...any) {
	if debugging {
		dbg.l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (dbg *Logger) Println(v ...any) {
	if debugging {
		dbg.l.Output(1, fmt.Sprint(v...))
	}
}
