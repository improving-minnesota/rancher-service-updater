package main

import (
	"testing"

	"github.com/rancher/go-rancher/client"
)

func Test_basic(t *testing.T) {
	_ = &ServiceUpdater{
		Config: &Config{
			EnableLabel:      "autoupdate.enable",
			EnvironmentNames: []string{".*"},
			Port:             8080,
		},
		service: &mockService{},
		account: &mockAccount{},
	}

}

type mockService struct{}
type mockAccount struct{}

func (a *mockService) ById(id string) (*client.Service, error) {
	return nil, nil
}

func (a *mockService) List(opts *client.ListOpts) (*client.ServiceCollection, error) {
	return nil, nil
}

func (a *mockService) ActionFinishupgrade(service *client.Service) (*client.Service, error) {
	return nil, nil
}

func (a *mockService) ActionUpgrade(service *client.Service, serviceUpgrade *client.ServiceUpgrade) (*client.Service, error) {
	return nil, nil
}

func (a *mockAccount) List(opts *client.ListOpts) (*client.AccountCollection, error) {
	return nil, nil
}
