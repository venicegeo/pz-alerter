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
