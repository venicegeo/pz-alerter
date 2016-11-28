// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syslog

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

//---------------------------------------------------------------------

// Logger is the "helper" class that can (should) be used by services to send messages.
// In most Piazza cases, the Writer field should be set to an ElkWriter.
type Logger struct {
	writer           WriterI
	MinimumSeverity  Severity // minimum severity level you want to record
	UseSourceElement bool
	application      string
	hostname         string
	processId        string
}

func NewLogger(writer WriterI, application string) *Logger {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNONW_HOSTNAME"
	}
	log.Printf("Hostname: %s", hostname)

	processId := strconv.Itoa(os.Getpid())

	logger := &Logger{
		writer:           writer,
		MinimumSeverity:  Informational,
		UseSourceElement: false,
		application:      application,
		hostname:         hostname,
		processId:        processId,
	}
	return logger
}

func (logger *Logger) severityAllowed(desiredSeverity Severity) bool {
	return logger.MinimumSeverity.Value() >= desiredSeverity.Value()
}

// postMessage sends a log message
func (logger *Logger) postMessage(severity Severity, text string, v ...interface{}) {
	if !logger.severityAllowed(severity) {
		return
	}

	newText := fmt.Sprintf(text, v...)

	mssg := NewMessage()
	mssg.Message = newText
	mssg.Severity = severity
	mssg.HostName = logger.hostname
	mssg.Application = logger.application
	mssg.Process = logger.processId

	if logger.UseSourceElement {
		// -1: stackFrame
		// 0: NewSourceElement
		// 1: postMessage
		// 2: Fatal
		// 3: actual source
		const skip = 3
		mssg.SourceData = NewSourceElement(skip)
	}

	err := logger.writer.Write(mssg)
	if err != nil {
		log.Printf("logger.postMessage: %s", err.Error())
	}
}

// Debug sends a log message with severity "Debug".
func (logger *Logger) Debug(text string, v ...interface{}) {
	logger.postMessage(Debug, text, v...)
}

// Info sends a log message with severity "Informational".
func (logger *Logger) Info(text string, v ...interface{}) {
	logger.postMessage(Informational, text, v...)
}

// Notice sends a log message with severity "Notice".
func (logger *Logger) Notice(text string, v ...interface{}) {
	logger.postMessage(Notice, text, v...)
}

// Warning sends a log message with severity "Warning".
func (logger *Logger) Warning(text string, v ...interface{}) {
	logger.postMessage(Warning, text, v...)
}

// Error sends a log message with severity "Error".
func (logger *Logger) Error(text string, v ...interface{}) {
	logger.postMessage(Error, text, v...)
}

// Fatal sends a log message with severity "Fatal".
func (logger *Logger) Fatal(text string, v ...interface{}) {
	logger.postMessage(Fatal, text, v...)
}
