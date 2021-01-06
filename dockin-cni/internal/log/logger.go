/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

var bizSeq string

func init() {
	loggerStderr = true
	loggerFp = nil
	loggerLevel = PanicLevel
	bizSeq = uuid.New().String()
}

type Level uint32

const (
	PanicLevel Level = iota
	ErrorLevel
	InfoLevel
	DebugLevel
	MaxLevel
	UnknownLevel
)

var loggerStderr bool
var loggerFp *os.File
var loggerLevel Level

const defaultTimestampFormat = time.RFC3339

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "panic"
	case InfoLevel:
		return "info"
	case ErrorLevel:
		return "error"
	case DebugLevel:
		return "debug"
	}
	return "unknown"
}

func printf(level Level, format string, a ...interface{}) {
	header := "%s [%s] "
	t := time.Now()
	if level > loggerLevel {
		return
	}

	funcName, _, line, _ := runtime.Caller(2)
	nf := fmt.Sprintf("-(%s:%d) %s, bizSeq=%%s", runtime.FuncForPC(funcName).Name(), line, format)
	a = append(a, bizSeq)

	if loggerStderr {
		fmt.Fprintf(os.Stderr, header, t.Format(defaultTimestampFormat), level)
		fmt.Fprintf(os.Stderr, nf, a...)
		fmt.Fprintf(os.Stderr, "\n")
	}

	if loggerFp != nil {
		fmt.Fprintf(loggerFp, header, t.Format(defaultTimestampFormat), level)
		fmt.Fprintf(loggerFp, nf, a...)
		fmt.Fprintf(loggerFp, "\n")
	}
}

func Debugf(format string, a ...interface{}) {
	printf(DebugLevel, format, a...)
}

func Infof(format string, a ...interface{}) {
	printf(InfoLevel, format, a...)
}

func Errorf(format string, a ...interface{}) error {
	printf(ErrorLevel, format, a...)
	return fmt.Errorf(format, a...)
}

func getloggerLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "verbose":
		return InfoLevel
	case "error":
		return ErrorLevel
	case "panic":
		return PanicLevel
	}
	fmt.Fprintf(os.Stderr, "logger: cannot set logger level to %s\n", levelStr)
	return UnknownLevel
}

func SetLogLevel(levelStr string) {
	level := getloggerLevel(levelStr)
	if level < MaxLevel {
		loggerLevel = level
	}
}

func SetLogFile(filename string) {
	if filename == "" {
		return
	}

	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		loggerFp = nil
		fmt.Fprintf(os.Stderr, "logger: cannot open %s", filename)
	}
	loggerFp = fp
}
