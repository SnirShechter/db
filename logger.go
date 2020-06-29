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
	"fmt"
	"log"
	"os"
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
	LogLevelWarn:  "WARNING",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
	LogLevelPanic: "PANIC",
}

const (
	defaultLogLevel LogLevel = LogLevelWarn
)

var defaultLogger Logger = log.New(os.Stdout, "upper/db: ", log.LstdFlags|log.Lmsgprefix)

// Logger
type Logger interface {
	Fatalf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}

// LoggingCollector represents a logging collector.
type LoggingCollector interface {
	SetLogger(Logger)
	SetLevel(LogLevel)

	Trace(format interface{}, v ...interface{})
	Debug(format interface{}, v ...interface{})
	Info(format interface{}, v ...interface{})
	Warn(format interface{}, v ...interface{})
	Error(format interface{}, v ...interface{})
	Fatal(format interface{}, v ...interface{})
	Panic(format interface{}, v ...interface{})
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

var defaultLoggingCollector LoggingCollector = &loggingCollector{
	level:  defaultLogLevel,
	logger: defaultLogger,
}

func Log() LoggingCollector {
	return defaultLoggingCollector
}

func init() {
	if logLevel := os.Getenv("UPPER_DB_LOG"); logLevel != "" {
		for k := range logLevels {
			if logLevels[k] == logLevel {
				Log().SetLevel(k)
				return
			}
		}
	}
}
