package client

import (
	"github.com/venicegeo/pz-gocommon"
)

type MockAlerterClient struct{}

func NewMockAlerterClient(sys *piazza.System) (*MockAlerterClient, error) {
	return &MockAlerterClient{}, nil
}

func (*MockAlerterClient) PostToEvents(*Event) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterClient) GetFromEvents() (*EventList, error) {
	return nil, nil
}

func (*MockAlerterClient) GetFromAlerts() (*AlertList, error) {
	return nil, nil
}

func (*MockAlerterClient) PostToConditions(*Condition) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerterClient) GetFromConditions() (*ConditionList, error) {
	return nil, nil
}

func (*MockAlerterClient) GetFromCondition(id string) (*Condition, error) {
	return nil, nil
}

func (*MockAlerterClient) DeleteOfCondition(id string) error {
	return nil
}

func (*MockAlerterClient) GetFromAdminStats() (*AlerterAdminStats, error) {
	return &AlerterAdminStats{}, nil
}

func (*MockAlerterClient) GetFromAdminSettings() (*AlerterAdminSettings, error) {
	return &AlerterAdminSettings{}, nil
}

func (*MockAlerterClient) PostToAdminSettings(*AlerterAdminSettings) error {
	return nil
}
