/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mock provides all methods required to simulate a driver
package mock

import (
	"net/http"
	"net/http/httptest"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

type MockTestEnv struct {
	Server *httptest.Server
	Mux    *http.ServeMux
	Client *hcloud.Client
}

func (env *MockTestEnv) Teardown() {
	env.Server.Close()

	env.Server = nil
	env.Mux = nil
	env.Client = nil
}

// CreateMachine handles a machine creation request
//
// PARAMETERS
// Machine      *v1alpha1.Machine      Machine object from whom VM is to be created
// MachineClass *v1alpha1.MachineClass MachineClass backing the machine object
// Secret       *corev1.Secret         Kubernetes secret that contains any sensitive data/credentials
//
func NewMockTestEnv() MockTestEnv {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	client := hcloud.NewClient(
		hcloud.WithEndpoint(server.URL),
		hcloud.WithHTTPClient(server.Client()),
	)

	return MockTestEnv{
		Server: server,
		Mux:    mux,
		Client: client,
	}
}
