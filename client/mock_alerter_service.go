package client

import (
	"github.com/venicegeo/pz-gocommon"
)

type MockAlerterService struct {
	name    string
	address string
}

func NewMockAlerterService(sys *piazza.System) (*MockAlerterService, error) {
	var _ IAlerterService = new(MockAlerterService)
	var _ piazza.IService = new(MockAlerterService)

	service := &MockAlerterService{name: piazza.PzAlerter, address: "0.0.0.0"}

	sys.Services[piazza.PzAlerter] = service

	return service, nil
}

func (m MockAlerterService) GetName() string {
	return m.name
}

func (m MockAlerterService) GetAddress() string {
	return m.address
}

func (*MockAlerterService) PostToEvents(*Event) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromEvents() (*EventList, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromAlerts() (*AlertList, error) {
	return nil, nil
}

func (*MockAlerterService) PostToConditions(*Condition) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromConditions() (*ConditionList, error) {
	return nil, nil
}

func (*MockAlerterService) GetFromCondition(id string) (*Condition, error) {
	return nil, nil
}

func (*MockAlerterService) DeleteOfCondition(id string) error {
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
