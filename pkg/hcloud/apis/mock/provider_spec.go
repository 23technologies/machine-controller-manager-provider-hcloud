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
	TestProviderSpecCluster = "xyz"
	TestProviderSpecDatacenter = "hel1-dc2"
	TestProviderSpecImageName = "ubuntu-20.04"
	TestProviderSpecServerType = "cx11-ceph"
	TestProviderSpecSSHFingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
	TestProviderSpec = "{\"cluster\":\"xyz\",\"datacenter\":\"hel1-dc2\",\"imageName\":\"ubuntu-20.04\",\"serverType\":\"cx11-ceph\",\"sshFingerprint\":\"00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff\"}"
	TestInvalidProviderSpec = "{\"test\":\"invalid\"}"
)

// ManipulateProviderSpec changes given provider specification.
//
// PARAMETERS
// providerSpec *apis.ProviderSpec      Provider specification
// data         map[string]interface{} Members to change
func ManipulateProviderSpec(providerSpec *apis.ProviderSpec, data map[string]interface{}) *apis.ProviderSpec {
	for key, value := range data {
		manipulateStruct(&providerSpec, key, value)
	}

	return providerSpec
}

// NewProviderSpec generates a new provider specification for testing purposes.
func NewProviderSpec() *apis.ProviderSpec {
	return &apis.ProviderSpec{
		Cluster: TestProviderSpecCluster,
		Datacenter: TestProviderSpecDatacenter,
		ImageName: TestProviderSpecImageName,
		ServerType: TestProviderSpecServerType,
		SSHFingerprint: TestProviderSpecSSHFingerprint,
	}
}
