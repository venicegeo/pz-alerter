package client

import (
)

type MockAlerter struct{}

func (*MockAlerter) PostToEvents(*Event) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerter) GetFromEvents() (*EventList, error) {
	return nil, nil
}

func (*MockAlerter) GetFromAlerts() (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerter) PostToConditions(*Condition) (*AlerterIdResponse, error) {
	return nil, nil
}

func (*MockAlerter) GetFromConditions() (*ConditionList, error) {
	return nil, nil
}

func (*MockAlerter) GetFromCondition(id string) (*Condition, error) {
	return nil, nil
}

func (*MockAlerter) DeleteOfCondition(id string) error {
	return nil
}

func (*MockAlerter) GetFromAdminStats() (*AlerterAdminStats, error) {
	return &AlerterAdminStats{}, nil
}

func (*MockAlerter) GetFromAdminSettings() (*AlerterAdminSettings, error) {
	return &AlerterAdminSettings{}, nil
}

func (*MockAlerter) PostToAdminSettings(*AlerterAdminSettings) error {
	return nil
}
