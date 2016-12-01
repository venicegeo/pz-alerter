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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bytes"

	"github.com/stretchr/testify/assert"
)

//---------------------------------------------------------------------

func fileEquals(t *testing.T, expected string, fileName string) {
	assert := assert.New(t)

	buf, err := ioutil.ReadFile(fileName)
	assert.NoError(err)

	assert.EqualValues(expected, string(buf))
}

func fileExist(s string) bool {
	if _, err := os.Stat(s); os.IsNotExist(err) {
		return false
	}
	return true
}

func safeRemove(s string) error {
	if fileExist(s) {
		err := os.Remove(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeMessage(sde bool) (*Message, string) {
	m := NewMessage()

	// because ParseMessageString isn't as accurate as we could be
	now := time.Now().Round(time.Second)

	m.Facility = DefaultFacility
	m.Severity = Fatal // pri = 1*8 + 2 = 10
	m.Version = DefaultVersion
	m.TimeStamp = now
	m.HostName = "HOST"
	m.Application = "APPLICATION"
	m.Process = "1234"
	m.MessageID = "msg1of2"
	m.AuditData = nil
	m.MetricData = nil
	m.Message = "Yow"

	expected := "<10>1 " + m.TimeStamp.Format(time.RFC3339) + " HOST APPLICATION 1234 msg1of2 - Yow"

	if sde {
		m.AuditData = &AuditElement{
			Actor:  "=actor=",
			Action: "-action-",
			Actee:  "_actee_",
		}
		m.MetricData = &MetricElement{
			Name:   "=name=",
			Value:  -3.14,
			Object: "_object_",
		}

		expected = "<10>1 " + m.TimeStamp.Format(time.RFC3339) + " HOST APPLICATION 1234 msg1of2 " +
			"[pzaudit@48851 actor=\"=actor=\" action=\"-action-\" actee=\"_actee_\"] " +
			"[pzmetric@48851 name=\"=name=\" value=\"-3.140000\" object=\"_object_\"] " +
			"Yow"
	}

	return m, expected
}

//---------------------------------------------------------------------

func Test01Message(t *testing.T) {
	assert := assert.New(t)

	m, expected := makeMessage(false)

	s := m.String()
	assert.EqualValues(expected, s)

	//	mm, err := ParseMessageString(expected)
	//	assert.NoError(err)

	//	assert.EqualValues(m, mm)
}

func Test02MessageSDE(t *testing.T) {
	assert := assert.New(t)

	m, expected := makeMessage(true)

	s := m.String()
	assert.EqualValues(expected, s)

	// TODO: this won't work until we make the parser understand SDEs
	//mm, err := ParseMessageString(expected)
	//assert.NoError(err)
	//assert.EqualValues(m, mm)
}

func Test03Writer(t *testing.T) {
	assert := assert.New(t)

	m, expected := makeMessage(false)

	{
		// verify error if no io.Writer given
		w := &Writer{Writer: nil}
		err := w.Write(m)
		assert.Error(err)
	}

	{
		// a simple kind of writer
		var buf bytes.Buffer
		w := &Writer{Writer: &buf}
		err := w.Write(m)
		assert.NoError(err)

		actual := buf.String()
		assert.EqualValues(expected, actual)
	}
}

func Test04FileWriter(t *testing.T) {
	var err error

	assert := assert.New(t)

	fname := "./testsyslog.txt"

	err = safeRemove(fname)
	assert.NoError(err)

	m1, expected1 := makeMessage(false)
	m2, expected2 := makeMessage(true)
	{
		w := &FileWriter{FileName: fname}
		err = w.Write(m1)
		assert.NoError(err)
		err = w.Close()
		assert.NoError(err)

		fileEquals(t, expected1+"\n", fname)
	}

	{
		w := &FileWriter{FileName: fname}
		err = w.Write(m2)
		assert.NoError(err)
		err = w.Close()
		assert.NoError(err)
		fileEquals(t, expected1+"\n"+expected2+"\n", fname)
	}

	err = safeRemove(fname)
	assert.NoError(err)
}

func Test05Logger(t *testing.T) {
	assert := assert.New(t)

	// the following clause is what a developer would do
	var buf bytes.Buffer
	{
		writer := &Writer{
			Writer: &buf,
		}
		logger := NewLogger(writer, "testapp")
		logger.UseSourceElement = false
		logger.Debug("debug %d", 999)
		logger.Info("info %d", 123)
		logger.Notice("notice %d", 321)
		logger.Warning("bonk %d", 3)
		logger.Error("Bonk %s", ".1")
		logger.Fatal("BONK %f", 4.0)
		logger.Audit("1", "2", "3", "brown%s", "fox")
		logger.Metric("i", 5952567, "k", "lazy%s", "dog")
	}

	mssg := buf.String()

	check := func(severity Severity, str string) {
		facility := DefaultFacility
		host, err := os.Hostname()
		assert.NoError(err)
		assert.Contains(mssg, fmt.Sprintf("<%d>", facility*8+severity.Value()))
		assert.Contains(mssg, fmt.Sprintf(" %d ", os.Getpid()))
		assert.Contains(mssg, fmt.Sprintf(" %s ", host))
		assert.Contains(mssg, str)
	}

	check(Debug, "debug 999")
	check(Informational, "info 123")
	check(Notice, "notice 321")
	check(Warning, "bonk 3")
	check(Error, "Bonk .1")
	check(Fatal, "BONK 4.0")
	check(Notice, "Actor=\"1\"")
	check(Notice, "brownfox")
	check(Notice, "Value=\"5952567.0")
	check(Notice, "lazydog")
}

func Test06LogLevel(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	{
		writer := &Writer{
			Writer: &buf,
		}
		logger := NewLogger(writer, "testapp")
		logger.UseSourceElement = false
		logger.MinimumSeverity = Error
		logger.Warning("bonk")
		logger.Error("Bonk")
		logger.Fatal("BONK")
	}

	mssg := buf.String()

	check := func(severity Severity, str string) {
		facility := DefaultFacility
		host, err := os.Hostname()
		assert.NoError(err)
		assert.Contains(mssg, fmt.Sprintf("<%d>", facility*8+severity.Value()))
		assert.Contains(mssg, fmt.Sprintf(" %d ", os.Getpid()))
		assert.Contains(mssg, fmt.Sprintf(" %s ", host))
		assert.Contains(mssg, str)
	}

	//pri(Warning, "bonk")
	assert.NotContains(mssg, "bonk")

	check(Error, "Bonk")
	check(Fatal, "BONK")
}

func Test07StackFrame(t *testing.T) {
	assert := assert.New(t)

	function, file, line := stackFrame(-1)
	//log.Printf("%s\t%s\t%d", function, file, line)
	assert.EqualValues(file, "SyslogMessage.go")
	assert.True(line > 1 && line < 1000)
	assert.EqualValues("syslog.stackFrame", function)

	function, file, line = stackFrame(0)
	//log.Printf("%s\t%s\t%d", function, file, line)
	assert.EqualValues(file, "Syslog_test.go")
	assert.True(line > 1 && line < 1000)
	assert.EqualValues("syslog.Test07StackFrame", function)
}
