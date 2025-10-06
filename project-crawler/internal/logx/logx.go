package logx

import (
	"log"
)

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Infow(msg string, kv ...any) {
	log.Println("INFO:", msg, kv)
}
func (l *Logger) Warnw(msg string, kv ...any) {
	log.Println("WARN:", msg, kv)
}
func (l *Logger) Errorw(msg string, kv ...any) {
	log.Println("ERROR:", msg, kv)
}
