package logx

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

// Logger wraps the standard logger and provides colored prefixes for levels.
type Logger struct {
	*log.Logger
	info *color.Color
	warn *color.Color
	err  *color.Color
}

// New returns a logger with colored level prefixes. We still use the
// standard library logger so timestamps and output configuration remain.
func New() *Logger {
	return &Logger{
		Logger: log.Default(),
		info:   color.New(color.FgWhite),
		warn:   color.New(color.FgYellow),
		err:    color.New(color.FgHiRed),
	}
}

func (l *Logger) Infow(msg string, kv ...interface{}) {
	// Create the message body and prepend a colored level prefix. Using
	// l.Printf keeps the standard timestamp/prefix behavior while the
	// colored prefix is embedded in the message.
	body := fmt.Sprintf("%s %v", msg, kv)
	l.Printf(l.info.Sprint("INFO:") + " " + body)
}

func (l *Logger) Warnw(msg string, kv ...interface{}) {
	body := fmt.Sprintf("%s %v", msg, kv)
	l.Printf(l.warn.Sprint("WARN:") + " " + body)
}

func (l *Logger) Errorw(msg string, kv ...interface{}) {
	body := fmt.Sprintf("%s %v", msg, kv)
	l.Printf(l.err.Sprint("ERROR:") + " " + body)
}
