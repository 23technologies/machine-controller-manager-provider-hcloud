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
	TestCluster = "xyz"
	TestImageName = "ubuntu-20.04"
	TestProviderSpec = "{\"cluster\":\"xyz\",\"zone\":\"hel1-dc2\",\"imageName\":\"ubuntu-20.04\",\"serverType\":\"cx11-ceph\",\"placementGroupID\":\"42\",\"sshFingerprint\":\"00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff\"}"
	TestServerType = "cx11-ceph"
	TestSSHFingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
	TestZone = "hel1-dc2"
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
		Cluster: TestCluster,
		Zone: TestZone,
		ImageName: TestImageName,
		ServerType: TestServerType,
		SSHFingerprint: TestSSHFingerprint,
		PlacementGroupID: TestPlacementGroupID,
	}
}
