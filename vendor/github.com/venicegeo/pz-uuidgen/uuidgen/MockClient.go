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
	"errors"
	"time"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

type MockClient struct {
	stats Stats
}

func NewMockClient() (*MockClient, error) {
	var _ IClient = new(MockClient)

	client := &MockClient{}

	client.stats.CreatedOn = time.Now()

	return client, nil
}

func (c *MockClient) GetVersion() (*piazza.Version, error) {
	version := piazza.Version{Version: Version}
	return &version, nil
}

func (c *MockClient) PostUuids(count int) (*[]string, error) {

	if count < 0 || count > 255 {
		return nil, errors.New("invalid count value")
	}

	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = piazza.NewUuid().String()
		//		data[i] = uuid.New()
	}

	c.stats.NumUUIDs += count
	c.stats.NumRequests++

	return &data, nil
}

func (c *MockClient) GetStats() (*Stats, error) {
	return &c.stats, nil
}

func (c *MockClient) GetUUID() (string, error) {
	data, err := c.PostUuids(1)
	if err != nil {
		return "", err
	}
	return (*data)[0], nil
}
