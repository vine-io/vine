// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/vine-io/vine/lib/dao/utils"
	log "github.com/vine-io/vine/lib/logger"
)

type LogLevel int

const (
	Silent LogLevel = iota + 1
	Error
	Warn
	Info
)

// Writer log writer interface
//type Writer interface {
//	Printf(string, ...interface{})
//}

type Options struct {
	SlowThreshold time.Duration
	LogLevel      LogLevel
}

// Interface logger interface
type Interface interface {
	Info(ctx context.Context, msg string, data ...interface{})
	Warn(ctx context.Context, msg string, data ...interface{})
	Error(ctx context.Context, msg string, data ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error)
}

var (
	Default = New(Options{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      Warn,
	})
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

func New(opt Options) Interface {
	return &logger{
		Helper:  log.NewHelper(log.DefaultLogger),
		Options: opt,
	}
}

type logger struct {
	*log.Helper
	Options
}

// LogMode log mode
func (l *logger) LogMode(level LogLevel) Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Infof(msg, data...)
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Warnf(msg, data...)
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Errorf(msg, data...)
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	fields := map[string]interface{}{
		"file": utils.FileWithLineNum(),
	}
	if l.LogLevel > Silent {
		elapsed := time.Since(begin)
		sql, rows := fc()
		fields["elapsed"] = float64(elapsed.Nanoseconds()) / 1e6
		switch {
		case err != nil && l.LogLevel >= Error:
			if rows == -1 {
				fields["rows"] = "-"
			} else {
				fields["rows"] = rows
			}
			l.Fields(fields).Log(log.ErrorLevel, "\n"+sql)
			l.Helper.Error(err)
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				fields["rows"] = "-"
			} else {
				fields["rows"] = rows
			}
			l.Fields(fields).Log(log.WarnLevel, "\n"+sql)
			l.Helper.Warn(slowLog)
		case l.LogLevel == Info:
			if rows == -1 {
				fields["rows"] = "-"
			} else {
				fields["rows"] = rows
			}
			l.Fields(fields).Log(log.InfoLevel, "\n"+sql)
		}
	}
}

type traceRecorder struct {
	Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

func (l traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: l.Interface, BeginAt: time.Now()}
}

func (l *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.BeginAt = begin
	l.SQL, l.RowsAffected = fc()
	l.Err = err
}
