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

type MockWorkflowService struct {
	name    piazza.ServiceName
	address string
}

func NewMockWorkflowService(sys *piazza.System) (*MockWorkflowService, error) {
	var _ IWorkflowService = new(MockWorkflowService)
	var _ piazza.IService = new(MockWorkflowService)

	service := &MockWorkflowService{name: piazza.PzWorkflow, address: "0.0.0.0"}

	sys.Services[piazza.PzWorkflow] = service

	return service, nil
}

func (m MockWorkflowService) GetName() piazza.ServiceName {
	return m.name
}

func (m MockWorkflowService) GetAddress() string {
	return m.address
}

func (*MockWorkflowService) PostToEvents(*Event) (*WorkflowIdResponse, error) {
	return nil, nil
}

func (*MockWorkflowService) GetFromEvents() (*[]Event, error) {
	return nil, nil
}

func (*MockWorkflowService) DeleteOfEvent(id Ident) error {
	return nil
}

func (*MockWorkflowService) GetFromAlerts() (*[]Alert, error) {
	return nil, nil
}

func (*MockWorkflowService) GetFromAlert(id Ident) (*Alert, error) {
	return nil, nil
}

func (*MockWorkflowService) PostToAlerts(*Alert) (*WorkflowIdResponse, error) {
	return nil, nil
}

func (*MockWorkflowService) DeleteOfAlert(id Ident) error {
	return nil
}

func (*MockWorkflowService) PostToTriggers(*Trigger) (*WorkflowIdResponse, error) {
	return nil, nil
}

func (*MockWorkflowService) GetFromTriggers() (*[]Trigger, error) {
	return nil, nil
}

func (*MockWorkflowService) GetFromTrigger(id Ident) (*Trigger, error) {
	return nil, nil
}

func (*MockWorkflowService) DeleteOfTrigger(id Ident) error {
	return nil
}

func (*MockWorkflowService) GetFromAdminStats() (*WorkflowAdminStats, error) {
	return &WorkflowAdminStats{}, nil
}

func (*MockWorkflowService) GetFromAdminSettings() (*WorkflowAdminSettings, error) {
	return &WorkflowAdminSettings{}, nil
}

func (*MockWorkflowService) PostToAdminSettings(*WorkflowAdminSettings) error {
	return nil
}
