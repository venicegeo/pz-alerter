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

package client

import (
	piazza "github.com/venicegeo/pz-gocommon"
)

type MockAlerterService struct {
	name    piazza.ServiceName
	address string
}

func NewMockAlerterService(sys *piazza.System) (*MockAlerterService, error) {
	var _ IAlerterService = new(MockAlerterService)
	var _ piazza.IService = new(MockAlerterService)

	service := &MockAlerterService{name: piazza.PzAlerter, address: "0.0.0.0"}

	sys.Services[piazza.PzAlerter] = service

	return service, nil
}

func (m MockAlerterService) GetName() piazza.ServiceName {
	return m.name
}

func (m MockAlerterService) GetAddress() string {
	return m.address
}

func (*MockAlerterService) PostToEvents(*Event) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromEvents() (*[]Event, error) {
	return nil, nil
}

func (*MockAlerterService) DeleteOfEvent(id Ident) error {
	return nil
}

func (*MockAlerterService) GetFromAlerts() (*[]Alert, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromAlert(id Ident) (*Alert, error) {
	return nil, nil
}

func (*MockAlerterService) PostToAlerts(*Alert) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterService) DeleteOfAlert(id Ident) error {
	return nil
}

func (*MockAlerterService) PostToTriggers(*Trigger) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromTriggers() (*[]Trigger, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromTrigger(id Ident) (*Trigger, error) {
	return nil, nil
}

func (*MockAlerterService) DeleteOfTrigger(id Ident) error {
	return nil
}

func (*MockAlerterService) GetFromAdminStats() (*AlerterAdminStats, error) {
	return &AlerterAdminStats{}, nil
}

func (*MockAlerterService) GetFromAdminSettings() (*AlerterAdminSettings, error) {
	return &AlerterAdminSettings{}, nil
}

func (*MockAlerterService) PostToAdminSettings(*AlerterAdminSettings) error {
	return nil
}
