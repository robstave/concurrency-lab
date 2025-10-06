package logx

import (
	"log"
)

type Logger struct{ *log.Logger }

func New() *Logger { return &Logger{log.Default()} }

func (l *Logger) Infow(msg string, kv ...interface{}) {
	l.Printf("INFO: %s %v", msg, kv)
}

func (l *Logger) Warnw(msg string, kv ...interface{}) {
	l.Printf("WARN: %s %v", msg, kv)
}

func (l *Logger) Errorw(msg string, kv ...interface{}) {
	l.Printf("ERROR: %s %v", msg, kv)
}
