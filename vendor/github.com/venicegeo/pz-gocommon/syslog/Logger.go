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
	"os"
	"strconv"
)

//---------------------------------------------------------------------

// Logger is the "helper" class that can (should) be used by services to send messages.
// Client needs to supply the right kind of Writer.
type Logger struct {
	writer           Writer
	MinimumSeverity  Severity // minimum severity level you want to record
	UseSourceElement bool
	application      string
	hostname         string
	processId        string
}

func NewLogger(writer Writer, application string) *Logger {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN_HOSTNAME"
	}

	processId := strconv.Itoa(os.Getpid())

	logger := &Logger{
		writer:           writer,
		MinimumSeverity:  Debug,
		UseSourceElement: true,
		application:      application,
		hostname:         hostname,
		processId:        processId,
	}

	return logger
}

func (logger *Logger) severityAllowed(desiredSeverity Severity) bool {
	return logger.MinimumSeverity.Value() >= desiredSeverity.Value()
}

// makeMessage sends a log message
func (logger *Logger) makeMessage(severity Severity, text string, v ...interface{}) *Message {

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

	return mssg
}

// postMessage sends a log message
func (logger *Logger) postMessage(mssg *Message) error {
	if logger.writer == nil {
		return fmt.Errorf("No writer set for logger")
	}

	if !logger.severityAllowed(mssg.Severity) {
		return nil
	}

	err := logger.writer.Write(mssg)
	if err != nil {
		return fmt.Errorf("logger.postMessage: %s <<%#v>>", err.Error(), mssg)
	}

	return nil
}

// Debug sends a log message with severity "Debug".
func (logger *Logger) Debug(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Debug, text, v...)
	return logger.postMessage(mssg)
}

// Information sends a log message with severity "Informational".
func (logger *Logger) Information(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Informational, text, v...)
	return logger.postMessage(mssg)
}

// Info is just an alternate name for Information
func (logger *Logger) Info(text string, v ...interface{}) error {
	return logger.Information(text, v...)
}

// Notice sends a log message with severity "Notice".
func (logger *Logger) Notice(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Notice, text, v...)
	return logger.postMessage(mssg)
}

// Warning sends a log message with severity "Warning".
func (logger *Logger) Warning(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Warning, text, v...)
	return logger.postMessage(mssg)
}

// Warn is just an alternate name for Warning
func (logger *Logger) Warn(text string, v ...interface{}) error {
	return logger.Warning(text, v...)
}

// Error sends a log message with severity "Error".
func (logger *Logger) Error(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Error, text, v...)
	return logger.postMessage(mssg)
}

// Fatal sends a log message with severity "Fatal".
func (logger *Logger) Fatal(text string, v ...interface{}) error {
	mssg := logger.makeMessage(Fatal, text, v...)
	return logger.postMessage(mssg)
}

// Audit sends a log message with the audit SDE.
func (logger *Logger) Audit(actor string, action string, actee string, text string, v ...interface{}) error {
	mssg := logger.makeMessage(Notice, text, v...)
	mssg.AuditData = &AuditElement{
		Actor:  actor,
		Action: action,
		Actee:  actee,
	}
	return logger.postMessage(mssg)
}

// Creates an audit message but does not send the message
func (logger *Logger) DebugAudit(actor string, action string, actee string, text string, v ...interface{}) error {
	mssg := logger.makeMessage(Notice, text, v...)
	mssg.AuditData = &AuditElement{
		Actor:  actor,
		Action: action,
		Actee:  actee,
	}
	return nil
	//return logger.postMessage(mssg)
}

// Metric sends a log message with the metric SDE.
func (logger *Logger) Metric(name string, value float64, object string, text string, v ...interface{}) error {
	mssg := logger.makeMessage(Notice, text, v...)
	mssg.MetricData = &MetricElement{
		Name:   name,
		Value:  value,
		Object: object,
	}
	return logger.postMessage(mssg)
}
