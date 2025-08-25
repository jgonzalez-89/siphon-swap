package logger

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type formatter struct {
	logrus.Formatter
	label  string
	level  string
	module string
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return fmt.Appendf([]byte{}, "[%s] [%s] [%s] [%s]: %s \n",
		entry.Time.Format(time.DateTime), f.level, f.label,
		f.module, entry.Message), nil
}
