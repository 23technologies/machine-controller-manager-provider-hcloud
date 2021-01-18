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
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
)

const (
	TestProviderSpecImageName = "ubuntu-20.04"
	TestProviderSpecServerType = "cx11-ceph"
	TestProviderSpecDatacenter = "hel1-dc2"
	TestProviderSpecKeyName = "test-ssh-publickey"
	TestProviderSpec = "{\"imageName\":\"ubuntu-20.04\",\"serverType\":\"cx11-ceph\",\"datacenter\":\"hel1-dc2\",\"keyName\":\"test-ssh-publickey\"}"
	TestInvalidProviderSpec = "{\"test\":\"invalid\"}"
)

// ManipulateProviderSpec changes given provider specification.
//
// PARAMETERS
// providerSpec *api.ProviderSpec      Provider specification
// data         map[string]interface{} Members to change
func ManipulateProviderSpec(providerSpec *api.ProviderSpec, data map[string]interface{}) *api.ProviderSpec {
	for key, value := range data {
		manipulateStruct(&providerSpec, key, value)
	}

	return providerSpec
}

// NewProviderSpec generates a new provider specification for testing purposes.
func NewProviderSpec() *api.ProviderSpec {
	return &api.ProviderSpec{
		ImageName: TestProviderSpecImageName,
		ServerType: TestProviderSpecServerType,
		Datacenter: TestProviderSpecDatacenter,
		KeyName: TestProviderSpecKeyName,
	}
}
