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

func simpleChecker(t *testing.T, m *Message, severity Severity, text string) {
	assert := assert.New(t)

	facility := DefaultFacility
	host, err := os.Hostname()
	assert.NoError(err)
	pid := fmt.Sprintf("%d", os.Getpid())

	assert.EqualValues(facility, m.Facility)
	assert.EqualValues(severity, m.Severity)
	assert.EqualValues(pid, m.Process)
	assert.EqualValues(host, m.HostName)
	assert.EqualValues(text, m.Message)
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

func Test03MessageWriter(t *testing.T) {
	assert := assert.New(t)

	mssg1, _ := makeMessage(false)
	mssg2, _ := makeMessage(false)

	w := &MessageWriter{}

	actual, err := w.Read(1)
	assert.NoError(err)
	assert.Len(actual, 0)

	err = w.Write(mssg1)
	assert.NoError(err)

	actual, err = w.Read(0)
	assert.NoError(err)
	assert.Len(actual, 0)

	actual, err = w.Read(1)
	assert.NoError(err)
	assert.Len(actual, 1)
	assert.EqualValues(mssg1, actual[0])

	actual, err = w.Read(2)
	assert.NoError(err)
	assert.Len(actual, 1)
	assert.EqualValues(mssg1, actual[0])

	err = w.Write(mssg2)
	assert.NoError(err)

	actual, err = w.Read(2)
	assert.NoError(err)
	assert.Len(actual, 2)
	assert.EqualValues(mssg1, actual[0])
	assert.EqualValues(mssg2, actual[1])

	actual, err = w.Read(-9)
	assert.Error(err)
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

	writer := &MessageWriter{}

	// the following clause is what a developer would do
	{
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

	mssgs, err := writer.Read(100)
	assert.NoError(err)
	assert.Len(mssgs, 8)

	simpleChecker(t, mssgs[0], Debug, "debug 999")
	simpleChecker(t, mssgs[1], Informational, "info 123")
	simpleChecker(t, mssgs[2], Notice, "notice 321")
	simpleChecker(t, mssgs[3], Warning, "bonk 3")
	simpleChecker(t, mssgs[4], Error, "Bonk .1")
	simpleChecker(t, mssgs[5], Fatal, "BONK 4.000000")
	simpleChecker(t, mssgs[6], Notice, "brownfox")
	assert.EqualValues("1", mssgs[6].AuditData.Actor)
	assert.EqualValues("2", mssgs[6].AuditData.Action)
	assert.EqualValues("3", mssgs[6].AuditData.Actee)
	simpleChecker(t, mssgs[7], Notice, "lazydog")
	assert.EqualValues("i", mssgs[7].MetricData.Name)
	assert.EqualValues(5952567, mssgs[7].MetricData.Value)
	assert.EqualValues("k", mssgs[7].MetricData.Object)
}

func Test06LogLevel(t *testing.T) {
	assert := assert.New(t)

	writer := &MessageWriter{}

	{
		logger := NewLogger(writer, "testapp")
		logger.UseSourceElement = false
		logger.MinimumSeverity = Error
		logger.Warning("bonk")
		logger.Error("Bonk")
		logger.Fatal("BONK")
	}

	mssgs, err := writer.Read(10)
	assert.NoError(err)
	assert.Len(mssgs, 2)

	simpleChecker(t, mssgs[0], Error, "Bonk")
	simpleChecker(t, mssgs[1], Fatal, "BONK")
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

func Test08Syslogd(t *testing.T) {
	assert := assert.New(t)

	m1, _ := makeMessage(false)

	w := &SyslogdWriter{}
	err := w.Write(m1)
	assert.NoError(err)
}
