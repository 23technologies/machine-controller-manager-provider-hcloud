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

// Package apis is the main package for provider specific APIs
package apis

// ProviderSpec is the spec to be used while parsing the calls.
type ProviderSpec struct {
	Cluster          string `json:"cluster"`
	Zone             string `json:"zone"`
	ServerType       string `json:"serverType"`
	ImageName        string `json:"imageName"`
	SSHFingerprint   string `json:"sshFingerprint"`

	PlacementGroupID string `json:"placementGroupID,omitempty"`
	FloatingPoolName string `json:"floatingPoolName,omitempty"`
	NetworkName      string `json:"networkName,omitempty"`
}
