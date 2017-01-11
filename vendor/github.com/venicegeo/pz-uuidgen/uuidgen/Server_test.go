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

package uuidgen

import (
	"fmt"
	"log"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
)

type UuidgenTester struct {
	suite.Suite
	sys            *piazza.SystemConfig
	totalRequested int
	totalGenerated int
	logWriter      pzsyslog.Writer
	auditWriter    pzsyslog.Writer
	client         IClient
	kit            *Kit
}

func (suite *UuidgenTester) SetupSuite() {
	var err error

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}

	suite.sys, err = piazza.NewSystemConfig(piazza.PzUuidgen, required)
	if err != nil {
		log.Fatal(err)
	}

	suite.logWriter = &pzsyslog.LocalReaderWriter{}
	suite.auditWriter = &pzsyslog.NilWriter{}

	suite.totalRequested = 0
	suite.totalGenerated = 0

	suite.kit, err = NewKit(suite.sys, suite.logWriter, suite.auditWriter)
	if err != nil {
		log.Fatal(err)
	}

	err = suite.kit.Start()
	if err != nil {
		log.Fatal(err)
	}

	suite.client, err = NewClient(suite.kit.Url, "")
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *UuidgenTester) TearDownSuite() {
	err := suite.kit.Stop()
	if err != nil {
		panic(err)
	}

	err = suite.logWriter.Close()
	if err != nil {
		panic(err)
	}
	err = suite.auditWriter.Close()
	if err != nil {
		panic(err)
	}
}

func TestRunSuite(t *testing.T) {
	s := new(UuidgenTester)
	suite.Run(t, s)
}

func (suite *UuidgenTester) checkValidStatsResponse(t *testing.T, stats *Stats) {
	assert.WithinDuration(t, time.Now(), stats.CreatedOn, 5*time.Second)

	assert.Equal(t, suite.totalGenerated, stats.NumUUIDs)
	assert.Equal(t, suite.totalRequested, stats.NumRequests)
}

func (suite *UuidgenTester) Test00Version() {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	version, err := client.GetVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("127.0.0.1:%s/version", piazza.LocalPortNumbers[piazza.PzUuidgen]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)
}

func (suite *UuidgenTester) Test01Okay() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	values := make([]string, 0)

	var client = suite.client

	data, err := client.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	values = append(values, *data...)
	suite.totalRequested++
	suite.totalGenerated++

	data, err = client.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	values = append(values, *data...)
	suite.totalRequested++
	suite.totalGenerated++

	data, err = client.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	values = append(values, *data...)
	suite.totalRequested++
	suite.totalGenerated++

	data, err = client.PostUuids(10)
	assert.NoError(err, "PostToUuids")
	values = append(values, *data...)
	suite.totalRequested++
	suite.totalGenerated += 10

	data, err = client.PostUuids(255)
	assert.NoError(err, "PostToUuids")
	values = append(values, *data...)
	suite.totalRequested++
	suite.totalGenerated += 255

	// uuids should be, umm, unique
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			p := values[j]
			q := values[i]
			if p == q {
				t.Fatalf("returned uuids not unique")
			}
		}
	}
	(*data)[0] = (*data)[1] // make compiler think data is actually used

	_, _, _, err = piazza.HTTP(piazza.POST, fmt.Sprintf("127.0.0.1:%s/uuids?count=0", piazza.LocalPortNumbers[piazza.PzUuidgen]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)
	suite.totalRequested++

	stats, err := client.GetStats()
	assert.NoError(err, "GetStats")
	suite.checkValidStatsResponse(t, stats)
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("127.0.0.1:%s/admin/stats", piazza.LocalPortNumbers[piazza.PzUuidgen]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)

	s, err := client.GetUUID()
	assert.NoError(err)
	assert.NotEmpty(s)
	suite.totalRequested++
	suite.totalGenerated++
}

func (suite *UuidgenTester) Test02Bad() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	var client = suite.client

	// count out of range
	_, err = client.PostUuids(-1)
	assert.Error(err)

	// count out of range
	_, err = client.PostUuids(256)
	assert.Error(err)
}
