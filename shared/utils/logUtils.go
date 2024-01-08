// SPDX-FileCopyrightText: 2023 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func LogInit(logToConsole bool) {
	zerolog.CallerMarshalFunc = logCallerMarshalFunction
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	fileWriter := getFileWriter()
	writers := []io.Writer{fileWriter}
	if logToConsole {
		writers = append(writers, zerolog.NewConsoleWriter())
	}

	multi := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Stack().Logger()
}

func getFileWriter() *lumberjack.Logger {
	const globalLogPath = "/var/log/"
	logPath := globalLogPath

	if file, err := os.OpenFile(globalLogPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600); err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			logPath = "./"
		}
		logPath = home
	} else {
		file.Close()
	}

	fileLogger := &lumberjack.Logger{
		Filename:   path.Join(logPath, "uyuni-tools.log"),
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     90,
		Compress:   true,
	}
	return fileLogger
}

func SetLogLevel(logLevel string) {
	globalLevel := zerolog.InfoLevel

	level, err := zerolog.ParseLevel(logLevel)
	if logLevel != "" && err == nil {
		globalLevel = level
	}
	if globalLevel <= zerolog.DebugLevel {
		log.Logger = log.Logger.With().Caller().Logger()
	}
	zerolog.SetGlobalLevel(globalLevel)
}

func logCallerMarshalFunction(pc uintptr, file string, line int) string {
	paths := strings.Split(file, "/")
	callerFile := file
	foundSubDir := false
	if strings.HasSuffix(file, "/io/io.go") {
		return "Cmd output"
	}

	for _, currentPath := range paths {
		if foundSubDir {
			if callerFile != "" {
				callerFile = callerFile + "/"
			}
			callerFile = callerFile + currentPath
		} else {
			if strings.Contains(currentPath, "uyuni-tools") {
				foundSubDir = true
				callerFile = ""
			}
		}
	}
	return callerFile + ":" + strconv.Itoa(line)
}
