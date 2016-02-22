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

package server

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-workflow/common"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"log"
	"testing"
	"time"
)

type ServerTester struct {
	suite.Suite
	logger     loggerPkg.ILoggerService
	uuidgenner uuidgenPkg.IUuidGenService
	sys        *piazza.System
}

func (suite *ServerTester) SetupSuite() {
	t := suite.T()
	assert := assert.New(t)

	config, err := piazza.NewConfig(piazza.PzWorkflow, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	suite.logger, err = loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	suite.uuidgenner, err = uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	routes, err := CreateHandlers(sys, suite.logger, suite.uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	suite.sys = sys

	assert.Len(sys.Services, 4)
}

func (suite *ServerTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(ServerTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) TestEventDB() {
	t := suite.T()
	assert := assert.New(t)

	es := suite.sys.ElasticSearchService

	et1 := common.Ident("T1")
	et2 := common.Ident("T2")

	var a1 common.Event
	a1.EventType = et1
	a1.Date = time.Now()

	var a2 common.Event
	a2.EventType = et2
	a2.Date = time.Now()

	db, err := NewEventDB(es, "event", "Event")
	assert.NoError(err)

	a1Id, err := db.PostData(&a1, NewResourceID())
	assert.NoError(err)
	a2Id, err := db.PostData(&a2, NewResourceID())
	assert.NoError(err)

	{
		raws, err := db.GetAll()
		assert.NoError(err)
		assert.Len(raws, 2)

		objs, err := ConvertRawsToEvents(raws)
		assert.NoError(err)

		ok1 := (objs[0].EventType == a1.EventType) && (objs[1].EventType == a2.EventType)
		ok2 := (objs[1].EventType == a1.EventType) && (objs[0].EventType == a2.EventType)
		assert.True((ok1 || ok2) && !(ok1 && ok2))
	}

	var t2 common.Event
	ok, err := db.GetById(a2Id, &t2)
	assert.NoError(err)
	assert.True(ok)
	assert.EqualValues(a2.EventType, t2.EventType)

	var t1 common.Event
	ok, err = db.GetById(a1Id, &t1)
	assert.NoError(err)
	assert.True(ok)
	assert.EqualValues(a1.EventType, t1.EventType)
}
