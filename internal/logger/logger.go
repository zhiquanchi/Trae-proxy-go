package logger

import (
	"log"
	"os"
)

// Logger 日志记录器
type Logger struct {
	debug bool
	info  *log.Logger
	err   *log.Logger
	debugLog *log.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(debug bool) *Logger {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	return &Logger{
		debug: debug,
		info:  log.New(os.Stdout, "[INFO] ", flags),
		err:   log.New(os.Stderr, "[ERROR] ", flags),
		debugLog: log.New(os.Stdout, "[DEBUG] ", flags),
	}
}

// Info 输出信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Error 输出错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debug {
		l.debugLog.Printf(format, v...)
	}
}

