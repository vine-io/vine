// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var deLogger Logger

func init() {
	DefaultOut(os.Stdout)
}

func DefaultOut(out io.Writer) {
	ws := zapcore.AddSync(out)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	deLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func Trace(args ...interface{}) {
	deLogger.Debug(args...)
}

func Tracef(format string, v ...interface{}) {
	deLogger.Debugf(format, v...)
}

func Debug(args ...interface{}) {
	deLogger.Debug(args...)
}

func Debugf(format string, v ...interface{}) {
	deLogger.Debugf(format, v...)
}

func Info(args ...interface{}) {
	deLogger.Info(args...)
}

func Infof(format string, v ...interface{}) {
	deLogger.Infof(format, v...)
}

func Warn(args ...interface{}) {
	deLogger.Warn(args...)
}

func Warnf(format string, v ...interface{}) {
	deLogger.Warnf(format, v...)
}

func Error(args ...interface{}) {
	deLogger.Error(args...)
}

func Errorf(format string, v ...interface{}) {
	deLogger.Errorf(format, v...)
}

func Fatal(args ...interface{}) {
	deLogger.Fatal(args...)
}

func Fatalf(format string, v ...interface{}) {
	deLogger.Fatalf(format, v...)
}
