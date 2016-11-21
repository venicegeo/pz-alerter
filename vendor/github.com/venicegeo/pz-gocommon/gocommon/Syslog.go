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

package piazza

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

//---------------------------------------------------------------------

// SyslogMessage represents all the fields of a native RFC5424 object, plus
// our own two SDEs.
type SyslogMessage struct {
	Facility    int            `json:"facility"`
	Severity    int            `json:"severity"`
	Version     int            `json:"version"`
	TimeStamp   string         `json:"timeStamp"`
	HostName    string         `json:"hostName"`
	IPAddress   string         `json:"ipAddress"`
	Application string         `json:"application"`
	Process     int            `json:"process"`
	MessageID   string         `json:"messageId"`
	AuditData   *AuditElement  `json:"auditData"`
	MetricData  *MetricElement `json:"metricData"`
	Message     string         `json:"message"`
}

// AuditElement represents an SDE for auditing (security-specific of just general).
type AuditElement struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Actee  string `json:"actee"`
}

// MetricElement represents an SDE for recoridng metrics.
type MetricElement struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Object string  `json:"object"`
}

// NewSyslogMessage returns a SyslogMessage with the defaults filled in for you.
func NewSyslogMessage() *SyslogMessage {
	m := &SyslogMessage{
		Facility: 1,
		// TODO: add all the other fields, with sensible defauls
		TimeStamp: time.Now().Format(time.RFC3339),
		Process:   os.Getpid(),
	}
	return m
}

// String builds and returns the RFC5424-style textual representation of a SyslogMessage.
func (m *SyslogMessage) String() string {
	// TODO: make this print all the fields, in the right format
	s := fmt.Sprintf("%d ... %s", m.Facility, m.Message)
	return s
}

func (m *SyslogMessage) validate() error {
	// TODO: add more/stronger error checks

	if m.HostName == "" && m.IPAddress == "" {
		return errors.New("Neither hostname nor IP address were supplied")
	} else if m.HostName != "" && m.IPAddress != "" {
		return errors.New("Both hostname and IP address were supplied")
	}
	if m.Facility != 1 {
		return errors.New("Bad facility")
	}
	if m.Severity != 2 && m.Severity != 3 && m.Severity != 4 && m.Severity != 5 && m.Severity != 6 && m.Severity != 7 {
		return errors.New("Bad severity")
	}
	if m.Version != 1 {
		return errors.New("Version is not 1")
	}

	_, err := time.Parse(time.RFC3339, m.TimeStamp)
	if err != nil {
		return errors.New("invalid time format")
	}

	if m.Application == "" {
		return errors.New("Application not set")
	}

	if m.Process == 0 {
		return errors.New("Process not set")
	}

	return nil
}

func (ae *AuditElement) validate() error {
	//TODO
	//Valid uuid
	//Valid action
	return nil
}

func (ae *MetricElement) validate() error {
	//TODO
	//Valid uuid
	//Valid action
	return nil
}

// Validate checks to see if a SyslogMessage is well-formed.
func (m *SyslogMessage) Validate() error {
	var err error

	err = m.validate()
	if err != nil {
		return err
	}

	if m.AuditData != nil {
		err = m.AuditData.validate()
		if err != nil {
			return err
		}
	}

	if m.MetricData != nil {
		err = m.MetricData.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

//---------------------------------------------------------------------

// SyslogWriter is an interface for writing a SyslogMessage to some sort of output.
type SyslogWriter interface {
	Write(*SyslogMessage) error
}

// SyslogSimpleWriter implements the SyslogWriter, writing to a generic "io.Writer" target
type SyslogSimpleWriter struct {
	Writer io.Writer
}

// Write writes the message to the io.Writer supplied.
func (w *SyslogSimpleWriter) Write(mssg *SyslogMessage) error {
	if w == nil || w.Writer == nil {
		return fmt.Errorf("writer not set not set")
	}

	s := mssg.String()
	_, err := io.WriteString(w.Writer, s)
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------

// SyslogFileWriter implements the SyslogWriter, writing to a given file
type SyslogFileWriter struct {
	FileName string
	file     *os.File
}

// Write writes the message to the supplied file.
func (w *SyslogFileWriter) Write(mssg *SyslogMessage) error {
	var err error

	if w == nil || w.FileName == "" {
		return fmt.Errorf("writer not set not set")
	}

	if w.file == nil {
		w.file, err = os.OpenFile(w.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
	}

	s := mssg.String()
	s += "\n"

	_, err = io.WriteString(w.file, s)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the file. The creator of the SyslogFileWriter must call this.
func (w *SyslogFileWriter) Close() error {
	return w.file.Close()
}

//---------------------------------------------------------------------

// Syslog is the "helper" class that can (should) be used by services to send messages.
// In most Piazza cases, the Writer field should be set to a SyslogElkWriter.
type Syslog struct {
	Writer SyslogWriter
}

// Warning sends a log message with severity "Warning".
func (syslog *Syslog) Warning(text string) {
	mssg := NewSyslogMessage()
	mssg.Message = text
	mssg.Severity = 123

	syslog.Writer.Write(mssg)
}

// Error sends a log message with severity "Error".
func (syslog *Syslog) Error(text string) {
	mssg := NewSyslogMessage()
	mssg.Message = text
	mssg.Severity = 345

	syslog.Writer.Write(mssg)
}

// Fatal sends a log message with severity "Fatal".
func (syslog *Syslog) Fatal(text string) {
	mssg := NewSyslogMessage()
	mssg.Message = text
	mssg.Severity = 567

	syslog.Writer.Write(mssg)
}
