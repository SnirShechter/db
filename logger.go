// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type LogLevel int8

const (
	LogLevelTrace LogLevel = -1

	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelPanic
)

var logLevels = map[LogLevel]string{
	LogLevelTrace: "TRACE",
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
	LogLevelPanic: "PANIC",
}

const (
	fmtLogSessID       = `Session ID:     %05d`
	fmtLogTxID         = `Transaction ID: %05d`
	fmtLogQuery        = `Query:          %s`
	fmtLogArgs         = `Arguments:      %#v`
	fmtLogRowsAffected = `Rows affected:  %d`
	fmtLogLastInsertID = `Last insert ID: %d`
	fmtLogError        = `Error:          %v`
	fmtLogTimeTaken    = `Time taken:     %0.5fs`
	fmtLogContext      = `Context:        %v`
)

var (
	reInvisibleChars = regexp.MustCompile(`[\s\r\n\t]+`)
)

// QueryStatus represents the status of a query after being executed.
type QueryStatus struct {
	SessID uint64
	TxID   uint64

	RowsAffected *int64
	LastInsertID *int64

	Query string
	Args  []interface{}

	Err error

	Start time.Time
	End   time.Time

	Context context.Context
}

// String returns a formatted log message.
func (q *QueryStatus) String() string {
	lines := make([]string, 0, 8)

	if q.SessID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogSessID, q.SessID))
	}

	if q.TxID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogTxID, q.TxID))
	}

	if query := q.Query; query != "" {
		query = reInvisibleChars.ReplaceAllString(query, ` `)
		query = strings.TrimSpace(query)
		lines = append(lines, fmt.Sprintf(fmtLogQuery, query))
	}

	if len(q.Args) > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogArgs, q.Args))
	}

	if q.RowsAffected != nil {
		lines = append(lines, fmt.Sprintf(fmtLogRowsAffected, *q.RowsAffected))
	}
	if q.LastInsertID != nil {
		lines = append(lines, fmt.Sprintf(fmtLogLastInsertID, *q.LastInsertID))
	}

	if q.Err != nil {
		lines = append(lines, fmt.Sprintf(fmtLogError, q.Err))
	}

	lines = append(lines, fmt.Sprintf(fmtLogTimeTaken, float64(q.End.UnixNano()-q.Start.UnixNano())/float64(1e9)))

	if q.Context != nil {
		lines = append(lines, fmt.Sprintf(fmtLogContext, q.Context))
	}

	return "\t" + strings.Replace(strings.Join(lines, "\n"), "\n", "\n\t", -1) + "\n\n"
}

const (
	defaultLogLevel = LogLevelWarn
)

var defaultLogger Logger = log.New(os.Stdout, "upper/db: ", log.LstdFlags|log.Lmsgprefix)

type Logger interface {
	Fatalf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}

// LoggingCollector represents a logging collector. You can pass a logging
// collector to db.DefaultSettings.SetLogger(myCollector) to make it collect
// db.QueryStatus messages after executing a query.
type LoggingCollector interface {
	SetLogger(Logger)
	Logger() Logger

	SetLevel(LogLevel)
	Level() LogLevel

	Trace(interface{}, ...interface{})
	Debug(interface{}, ...interface{})
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{})
	Error(interface{}, ...interface{})
	Fatal(interface{}, ...interface{})
	Panic(interface{}, ...interface{})
}

type loggingCollector struct {
	level  LogLevel
	logger Logger
}

func (c *loggingCollector) SetLevel(level LogLevel) {
	c.level = level
}

func (c *loggingCollector) Level() LogLevel {
	return c.level
}

func (c *loggingCollector) Logger() Logger {
	if c.logger == nil {
		return defaultLogger
	}
	return c.logger
}

func (c *loggingCollector) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *loggingCollector) log(level LogLevel, f interface{}, v ...interface{}) {
	if level < c.level {
		return
	}
	format := logLevels[c.level] + "\n" + fmt.Sprintf("%v", f)

	if c.level >= LogLevelPanic {
		c.Logger().Panicf(format, v...)
	}
	if c.level >= LogLevelFatal {
		c.Logger().Fatalf(format, v...)
	}
	c.Logger().Printf(format, v...)
}

func (c *loggingCollector) Debug(format interface{}, v ...interface{}) {
	c.log(LogLevelDebug, format, v...)
}

func (c *loggingCollector) Trace(format interface{}, v ...interface{}) {
	c.log(LogLevelTrace, format, v...)
}

func (c *loggingCollector) Info(format interface{}, v ...interface{}) {
	c.log(LogLevelInfo, format, v...)
}

func (c *loggingCollector) Warn(format interface{}, v ...interface{}) {
	c.log(LogLevelWarn, format, v...)
}

func (c *loggingCollector) Error(format interface{}, v ...interface{}) {
	c.log(LogLevelError, format, v...)
}

func (c *loggingCollector) Fatal(format interface{}, v ...interface{}) {
	c.log(LogLevelFatal, format, v...)
}

func (c *loggingCollector) Panic(format interface{}, v ...interface{}) {
	c.log(LogLevelPanic, format, v...)
}

var defaultLoggingCollector = &loggingCollector{
	level:  defaultLogLevel,
	logger: defaultLogger,
}

func Log() LoggingCollector {
	return defaultLoggingCollector
}
