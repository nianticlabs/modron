package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// setLogLevel sets the logrus log level
func setLogLevel() {
	switch strings.ToLower(args.LogLevel) {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warning":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	}
}

type gcpFormatter struct{}

func (g gcpFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := map[string]any{}
	output["severity"] = toGcpSeverity(entry.Level)
	output["message"] = entry.Message
	output["timestamp"] = entry.Time.Format(time.RFC3339Nano)
	if len(entry.Data) > 0 {
		output["labels"] = entry.Data
	}

	b, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, nil
}

func toGcpSeverity(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel, logrus.DebugLevel:
		return "DEBUG"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARNING"
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.FatalLevel, logrus.PanicLevel:
		return "CRITICAL"
	default:
		return "DEFAULT"
	}
}

var gcpFormat = &gcpFormatter{}
var textFormatter = &logrus.TextFormatter{}

// setLogFormat sets the format of the logs
func setLogFormat() {
	var formatter logrus.Formatter
	invalidFormatter := false
	switch args.LogFormat {
	case LogFormatJSON:
		formatter = gcpFormat
	case LogFormatText:
		formatter = textFormatter
	default:
		formatter = textFormatter
		invalidFormatter = true
	}

	logrus.SetFormatter(formatter)
	if invalidFormatter {
		log.Errorf("invalid log format, using %s", LogFormatText)
	}
}
