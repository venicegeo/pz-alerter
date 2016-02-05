package client

import (
	"github.com/venicegeo/pz-gocommon"
)

type MockAlerterService struct{
	Name string
	Address string
}

func NewMockAlerterService(sys *piazza.System) (*MockAlerterService, error) {
	var _ IAlerterService = new(MockAlerterService)
	var _ piazza.IService = new(MockAlerterService)

	return &MockAlerterService{Name: "pz-alerter", Address: "0.0.0.0"}, nil
}

func (m *MockAlerterService) GetName() string {
	return m.Name
}

func (m *MockAlerterService) GetAddress() string {
	return m.Address
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
