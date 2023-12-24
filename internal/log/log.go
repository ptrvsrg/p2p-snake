package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type CustomTextFormatter struct{}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := fmt.Sprintf("%s %5s --- %s:%d : %s \n",
		entry.Time.Format("2006-01-02 15:04:05.000"),
		strings.ToUpper(entry.Level.String()),
		entry.Caller.File,
		entry.Caller.Line,
		entry.Message,
	)

	return []byte(message), nil
}

var Logger = &logrus.Logger{
	Out:          os.Stdout,
	Level:        logrus.InfoLevel,
	Formatter:    &CustomTextFormatter{},
	ReportCaller: true,
}
