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

package workflow

import (
	"encoding/json"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

// CronDB TODO
type CronDB struct {
	*ResourceDB
	mapping string
}

// NewCronDB TODO
func NewCronDB(service *Service, esi elasticsearch.IIndex) (*CronDB, error) {
	rdb, err := NewResourceDB(service, esi, CronIndexSettings)
	if err != nil {
		return nil, err
	}
	crdb := CronDB{ResourceDB: rdb, mapping: cronDBMapping}
	return &crdb, nil
}

// PostData TODO
func (db *CronDB) PostData(obj interface{}, id piazza.Ident, actor string) error {
	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		db.service.syslogger.Audit(actor, "createCron", string(id), "CronDB.PostData: failed")
		return LoggedError("CronDB.PostData failed: %s", err)
	} else if !indexResult.Created {
		db.service.syslogger.Audit(actor, "createCron", string(id), "CronDB.PostData: failed")
		return LoggedError("CronDB.PostData failed: not created")
	}
	db.service.syslogger.Audit(actor, "createCron", string(id), "CronDB.PostData: success")

	return nil
}

// GetAll TODO
func (db *CronDB) GetAll(actor string) (*[]Event, error) {
	var events []Event

	db.service.syslogger.Audit(actor, "readType", db.mapping, "CronDB.GetAll: check type exists")
	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return &events, err
	}
	if !exists {
		return nil, LoggedError("Type %s does not exist", db.mapping)
	}

	db.service.syslogger.Audit(actor, "readCrons", db.mapping, "CronDB.GetAll: get all query")
	searchResult, err := db.Esi.GetAllElements(db.mapping)
	if err != nil {
		return nil, LoggedError("CronDB.GetAll failed: %s", err)
	} else if searchResult == nil {
		return nil, LoggedError("CronDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var event Event
			err := json.Unmarshal(*hit.Source, &event)
			if err != nil {
				return nil, LoggedError("CronDB.GetAll failed: %s", err)
			}
			events = append(events, event)
		}
	}

	return &events, nil
}

// Exists checks to see if the database exists
func (db *CronDB) Exists(actor string) (bool, error) {
	db.service.syslogger.Audit(actor, "readIndex", db.Esi.IndexName(), "CronDB.Exists: check index exists")
	exists, err := db.Esi.IndexExists()
	if err != nil {
		return false, err
	}
	if exists {
		db.service.syslogger.Audit(actor, "readType", db.mapping, "CronDB.Exists: check type exists")
		exists, err = db.Esi.TypeExists(db.mapping)
		if err != nil {
			return false, err
		}
	}
	return exists, nil
}

func (db *CronDB) itemExists(id piazza.Ident, actor string) (bool, error) {
	db.service.syslogger.Audit(actor, "readCron", string(id), "CronDB.itemExists: check item exists")
	return db.Esi.ItemExists(db.mapping, id.String())
}

func (db *CronDB) DeleteByID(id piazza.Ident, actor string) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		db.service.syslogger.Audit(actor, "deleteCron", string(id), "CronDB.DeleteByID: failed")
		return deleteResult.Found, LoggedError("CronDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		db.service.syslogger.Audit(actor, "deleteCron", string(id), "CronDB.DeleteByID: failed")
		return false, LoggedError("CronDB.DeleteById failed: no deleteResult")
	}
	db.service.syslogger.Audit(actor, "deleteCron", string(id), "CronDB.DeleteByID: sucess")

	return deleteResult.Found, nil
}
