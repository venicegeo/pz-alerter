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

import "github.com/venicegeo/pz-gocommon/elasticsearch"

type ResourceDB struct {
	service *Service
	Esi     elasticsearch.IIndex
}

func NewResourceDB(service *Service, esi elasticsearch.IIndex, settings string) (*ResourceDB, error) {
	db := &ResourceDB{
		service: service,
		Esi:     esi,
	}

	// _ = esi.Delete()
	err := esi.Create(settings)
	if err != nil {
		db.service.syslogger.Audit("pz-workflow", "createIndex", esi.IndexName(), "NewResourceDB: failed")
		return nil, err
	}
	db.service.syslogger.Audit("pz-workflow", "createIndex", esi.IndexName(), "NewResourceDB: success")

	return db, nil
}
