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

package workflow_systest

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-workflow/workflow"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type WorkflowTester struct {
	suite.Suite
	client *workflow.Client
	url    string
	apiKey string
}

func (suite *WorkflowTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	suite.url = "https://pz-workflow.int.geointservices.io"

	suite.apiKey, err = piazza.GetApiKey("int")
	assert.NoError(err)

	client, err := workflow.NewClient2(suite.url, suite.apiKey)
	assert.NoError(err)
	suite.client = client
}

func (suite *WorkflowTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &WorkflowTester{}
	suite.Run(t, s)
}

func (suite *WorkflowTester) TestGet() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	eventTypes, err := client.GetAllEventTypes()
	assert.NoError(err)

	assert.True(len(*eventTypes) > 1)
}

func (suite *WorkflowTester) TestPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	uniq := "systest$" + strconv.Itoa(time.Now().Nanosecond())

	eventType := &workflow.EventType{
		Name: uniq,
		Mapping: map[string]elasticsearch.MappingElementTypeName{
			"filename": elasticsearch.MappingElementTypeString,
			"code":     elasticsearch.MappingElementTypeString,
			"severity": elasticsearch.MappingElementTypeInteger,
		},
	}

	ack, err := client.PostEventType(eventType)
	assert.NoError(err)
	assert.NotNil(ack)
}

func (suite *WorkflowTester) TestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	stats, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")

	assert.NotZero(stats.NumEvents)
}
