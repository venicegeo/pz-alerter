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

package uuidgen_systest

import (
	"testing"
	"time"

	uuidpkg "github.com/pborman/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-uuidgen/uuidgen"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type UuidgenTester struct {
	suite.Suite
	client *uuidgen.Client
	url    string
	apiKey string
}

func (suite *UuidgenTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	suite.url = "https://pz-uuidgen.int.geointservices.io"

	suite.apiKey, err = piazza.GetApiKey("int")
	assert.NoError(err)

	client, err := uuidgen.NewClient2(suite.url, suite.apiKey)
	assert.NoError(err)
	suite.client = client
}

func (suite *UuidgenTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &UuidgenTester{}
	suite.Run(t, s)
}

func isValid(uuid string) bool {
	return uuidpkg.Parse(uuid) != nil
}

func (suite *UuidgenTester) TestGet() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	uuid, err := client.GetUuid()
	assert.NoError(err)

	assert.True(isValid(uuid))
}

func (suite *UuidgenTester) TestPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	uuids, err := client.PostUuids(17)
	assert.NoError(err)
	assert.Len(*uuids, 17)

	for i := 0; i < 17; i++ {
		a := (*uuids)[i]
		assert.True(isValid(a))
		for j := i + 1; j < 17; j++ {
			b := (*uuids)[j]
			assert.NotEqual(a, b)
		}
	}
}

func (suite *UuidgenTester) TestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	stats, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")

	assert.NotZero(stats.NumUUIDs)
	assert.NotZero(stats.NumRequests)
}
