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

package main

import (
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	pzlogger "github.com/venicegeo/pz-logger/logger"
)

func main() {

	required := []piazza.ServiceName{piazza.PzElasticSearch}

	sys, err := piazza.NewSystemConfig(piazza.PzLogger, required)
	if err != nil {
		log.Fatal(err)
	}

	idx, logESWriter, err := setupES(sys)
	if err != nil {
		log.Fatal(err)
	}

	stdoutWriter := pzsyslog.StdoutWriter{}

	kit, err := pzlogger.NewKit(sys, logESWriter, &stdoutWriter, idx, true)
	if err != nil {
		log.Fatal(err)
	}

	err = kit.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = kit.Wait()
	if err != nil {
		log.Fatal(err)
	}

	err = closeES(idx, logESWriter)
	if err != nil {
		log.Fatal(err)
	}
}

func closeES(idx elasticsearch.IIndex, logWriter pzsyslog.Writer) error {
	err := logWriter.Close()
	if err != nil {
		log.Fatal(err)
	}

	return idx.Close()
}

func setupES(sys *piazza.SystemConfig) (elasticsearch.IIndex, pzsyslog.Writer, error) {
	loggerIndex, loggerType, err := pzsyslog.GetRequiredEnvVars()
	if err != nil {
		log.Fatal(err)
	}
	pzlogger.SetLogSchema(loggerType)

	idx, err := elasticsearch.NewIndex(sys, loggerIndex, "")
	if err != nil {
		return nil, nil, err
	}

	logESWriter, err := pzsyslog.GetRequiredESIWriters(idx, loggerType)
	if err != nil {
		return nil, nil, err
	}

	return idx, logESWriter, nil
}
