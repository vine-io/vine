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
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func Default() {
	ws := zapcore.AddSync(os.Stdout)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func Debug(s string) {
	sugar.Debugw(s)
}

func Debugf(format string, v ...interface{}) {
	sugar.Debugf(format, v...)
}

func Info(s string) {
	sugar.Infow(s)
}

func Infof(format string, v ...interface{}) {
	sugar.Infof(format, v...)
}

func Warn(s string) {
	sugar.Warnw(s)
}

func Warnf(format string, v ...interface{}) {
	sugar.Warnf(format, v...)
}

func Error(s string) {
	sugar.Errorw(s)
}

func Errorf(format string, v ...interface{}) {
	sugar.Errorf(format, v...)
}

func Fatal(s string) {
	sugar.Fatalw(s)
}

func Fatalf(format string, v ...interface{}) {
	sugar.Fatalf(format, v...)
}

func Sync() error {
	return sugar.Sync()
}
