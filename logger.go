package context

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type LoggerFormatMessage struct {
	Header *Header                `json:"header"`
	Info   any                    `json:"info,omitempty"`
	Error  string                 `json:"error,omitempty"`
	Stack  string                 `json:"stack,omitempty"`
	Extra  map[string]interface{} `json:"extra,omitempty"`
}

type LoggerMessage struct {
	Header   *Header `json:"header"`
	Location string  `json:"-"`
	Message  string  `json:"-"`
	Data     any     `json:"data,omitempty"`
	Error    string  `json:"error,omitempty"`
	Stack    string  `json:"stack,omitempty"`
}

func (m *LoggerMessage) MakeLoggerFormatMessage() (formatMessage *LoggerFormatMessage) {
	return &LoggerFormatMessage{
		Header: m.Header,
		Info:   m.Data,
		Error:  m.Error,
		Stack:  m.Stack,
	}
}

func (m *LoggerMessage) Fields() logrus.Fields {
	return logrus.Fields{loggerMessageKey: m}
}

type ILogger interface {
	Info(message *LoggerMessage)
	Trace(message *LoggerMessage)
	Debug(message *LoggerMessage)
	Warn(message *LoggerMessage)
	Fatal(message *LoggerMessage)
	Panic(message *LoggerMessage)
	Error(message *LoggerMessage, err error)
}

func NewLogger(name string, opt *LoggerConfig) *Logger {
	return &Logger{
		entry: NewLoggerBuilder(name, opt).Make(),
		opt:   opt,
	}
}

func DefaultLogger(name string) *Logger {
	return NewLogger(name, defaultLoggerConfig)
}

type Logger struct {
	entry *logrus.Logger
	opt   *LoggerConfig
}

func (l Logger) recover() {
	if r := recover(); r != nil {
		l.Error(&LoggerMessage{
			Message: "log failed",
			Error:   fmt.Sprintf("%+v", r),
			Stack:   takeStack(),
		}, nil)
	}
}

func (l Logger) asyncLog(level logrus.Level, message *LoggerMessage) {
	defer l.recover()
	l.entry.WithFields(message.Fields()).Log(level)
}

func (l Logger) syncLog(level logrus.Level, message *LoggerMessage) {
	l.entry.WithFields(message.Fields()).Log(level)
}

func (l Logger) Log(level logrus.Level, message *LoggerMessage) {
	if l.opt.Async {
		go l.asyncLog(level, message)
		return
	}

	l.syncLog(level, message)
}

func (l Logger) Info(message *LoggerMessage) {
	l.Log(logrus.InfoLevel, message)
}

func (l Logger) Trace(message *LoggerMessage) {
	l.Log(logrus.TraceLevel, message)
}

func (l Logger) Debug(message *LoggerMessage) {
	l.Log(logrus.DebugLevel, message)
}

func (l Logger) Warn(message *LoggerMessage) {
	l.Log(logrus.WarnLevel, message)
}

func (l Logger) Fatal(message *LoggerMessage) {
	l.Log(logrus.FatalLevel, message)
}

func (l Logger) Panic(message *LoggerMessage) {
	l.Log(logrus.PanicLevel, message)
}

func (l Logger) Error(message *LoggerMessage, err error) {
	if err != nil {
		message.Error = err.Error()
		message.Stack = takeStack()
	}

	l.Log(logrus.ErrorLevel, message)
}
