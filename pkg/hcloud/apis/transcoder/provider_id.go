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

// Package transcoder is used for API related object transformations
package transcoder

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"strconv"
)

// ProviderSpec is the spec to be used while parsing the calls.
type ServerData struct {
	ID   int
	Zone string
}

// DecodeServerDataFromProviderID decodes the given provider ID to extract the server specific data.
//
// PARAMETERS
// providerID string Provider ID to parse
func DecodeServerDataFromProviderID(providerID string) (*ServerData, error) {
	providerIDUrl, err := url.Parse(providerID)
	if err != nil {
		return nil, fmt.Errorf("ProviderID given is malformed: %v", err)
	} else if providerIDUrl.Scheme != "hcloud" {
		return nil, errors.New("ProviderID given contains an unsupported URL scheme")
	}

	providerIDData := strings.SplitN(providerIDUrl.Path[1:], "/", 2)
	if len(providerIDData) != 2 {
		return nil, errors.New("ProviderID given contains an invalid URL")
	}

	serverID, err := strconv.Atoi(providerIDData[1])
	if err != nil {
		return nil, fmt.Errorf("ServerID found is invalid: %v", err)
	}

	response := &ServerData{
		ID:   serverID,
		Zone: providerIDData[0],
	}

	return response, nil
}

// DecodeZoneFromProviderID decodes the given ProviderID to extract the datacenter zone.
//
// PARAMETERS
// providerID string Provider ID to parse
func DecodeZoneFromProviderID(providerID string) (string, error) {
	serverData, err := DecodeServerDataFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	return serverData.Zone, nil
}

// DecodeServerIDFromProviderID decodes the given ProviderID to extract the server ID.
//
// PARAMETERS
// providerID string Provider ID to parse
func DecodeServerIDFromProviderID(providerID string) (int, error) {
	serverData, err := DecodeServerDataFromProviderID(providerID)
	if err != nil {
		return -1, err
	}

	return serverData.ID, nil
}

// DecodeServerIDAsStringFromProviderID decodes the given ProviderID to extract the server ID.
//
// PARAMETERS
// providerID string Provider ID to parse
func DecodeServerIDAsStringFromProviderID(providerID string) (string, error) {
	serverData, err := DecodeServerDataFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(serverData.ID), nil
}

// EncodeProviderID encodes the ProviderID string based on the given zone and server ID.
//
// PARAMETERS
// providerID string Provider ID to parse
func EncodeProviderID(zone string, serverID int) string {
	return fmt.Sprintf("hcloud:///%s/%d", url.PathEscape(zone), serverID)
}
