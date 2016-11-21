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
	"io/ioutil"
	"os"
	"testing"

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

//---------------------------------------------------------------------

func Test01SyslogMessage(t *testing.T) {
	assert := assert.New(t)

	m := &SyslogMessage{
		Facility: 99,
		Message:  "Yow!",
	}
	s := m.String()
	assert.EqualValues("99 ... Yow!", s)
}

func Test02SyslogWriter(t *testing.T) {
	var err error

	assert := assert.New(t)

	fname := "./tmptmp"

	m := &SyslogMessage{
		Facility: 99,
		Message:  "Yow!",
	}

	w := &SyslogSimpleWriter{Writer: nil}

	err = safeRemove(fname)
	assert.NoError(err)

	// verify error if no io.Writer given
	err = w.Write(m)
	assert.Error(err)

	f, err := os.Create(fname)
	assert.NoError(err)
	defer f.Close()
	w.Writer = f

	err = w.Write(m)
	assert.NoError(err)

	err = safeRemove(fname)
	assert.NoError(err)
}

func Test03SyslogFileWriter(t *testing.T) {
	var err error

	assert := assert.New(t)

	fname := "./tmptmp2"

	err = safeRemove(fname)
	assert.NoError(err)

	{
		m := &SyslogMessage{
			Facility: 99,
			Message:  "One!",
		}
		w := &SyslogFileWriter{FileName: fname}
		err = w.Write(m)
		assert.NoError(err)
		err = w.Close()
		assert.NoError(err)
		fileEquals(t, "99 ... One!\n", fname)
	}

	{
		m := &SyslogMessage{
			Facility: 100,
			Message:  "Two!",
		}
		w := &SyslogFileWriter{FileName: fname}
		err = w.Write(m)
		assert.NoError(err)
		err = w.Close()
		assert.NoError(err)
		fileEquals(t, "99 ... One!\n100 ... Two!\n", fname)
	}

	err = safeRemove(fname)
	assert.NoError(err)
}

func Test04Syslog(t *testing.T) {
	var err error

	assert := assert.New(t)

	logfile := "./mylog.txt"

	err = safeRemove(logfile)
	assert.NoError(err)

	// this is what a developer would do
	{
		writer := &SyslogFileWriter{
			FileName: logfile,
		}
		logger := &Syslog{
			Writer: writer,
		}
		logger.Warning("bonk")
		logger.Error("Bonk")
		logger.Fatal("BONK")
	}

	fileEquals(t, "1 ... bonk\n1 ... Bonk\n1 ... BONK\n", logfile)

	err = safeRemove(logfile)
	assert.NoError(err)
}
