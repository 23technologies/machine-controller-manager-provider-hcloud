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
	"fmt"
	"encoding/json"
	"net/http"
	"strings"

	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	jsonImageData = `
{
	"id": 42,
	"type": "snapshot",
	"status": "available",
	"name": "ubuntu-20.04",
	"description": "Proudly copied from the Hetzner Cloud API documentation",
	"image_size": 2.3,
	"disk_size": 10,
	"created": "2016-01-30T23:50:00+00:00",
	"created_from": {
		"id": 1,
		"name": "Server"
	},
	"os_flavor": "ubuntu",
	"os_version": "20.04",
	"rapid_deploy": false,
	"protection": {
		"delete": false
	},
	"deprecated": "2018-02-28T00:00:00+00:00",
	"labels": {}
}
	`
	TestNamespace = "test"
)

func NewMachine(setMachineIndex int) *v1alpha1.Machine {
	index := 0

	if setMachineIndex > 0 {
		index = setMachineIndex
	}

	machine := &v1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "machine.sapcloud.io",
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("machine-%d", index),
			Namespace: TestNamespace,
		},
	}

	// Don't initialize providerID and node if setMachineIndex == -1
	if setMachineIndex != -1 {
		machine.Spec = v1alpha1.MachineSpec{
			ProviderID: fmt.Sprintf("hcloud:///%s/i-0123456789-%d", TestProviderSpecDatacenter, setMachineIndex),
		}
		machine.Status = v1alpha1.MachineStatus{
			Node: fmt.Sprintf("ip-%d", setMachineIndex),
		}
	}

	return machine
}

func NewMachineClass() *v1alpha1.MachineClass {
	return NewMachineClassWithProviderSpec([]byte(TestProviderSpec))
}

func NewMachineClassWithProviderSpec(providerSpec []byte) *v1alpha1.MachineClass {
	return &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: providerSpec,
		},
	}
}

func SetupImagesEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/images", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		res.Write([]byte(`
{
	"images": [
		`))

		queryParams := req.URL.Query()

		if (queryParams.Get("name") == TestProviderSpecImageName) {
			res.Write([]byte(jsonImageData))
		}

		res.Write([]byte(`
	]
}
		`))
	})
}

func SetupServersEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/servers", func(res http.ResponseWriter, req *http.Request) {
		var (
			data map[string]interface{}
		)

		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusCreated)

			jsonData := make([]byte, req.ContentLength)
			req.Body.Read(jsonData)

			jsonErr := json.Unmarshal(jsonData, &data)
			if jsonErr != nil {
				panic(jsonErr)
			}

			res.Write([]byte(fmt.Sprintf(`
{
	"server": {
		"id": 42,
		"name": "%s",
		"status": "initializing",
		"created": "2016-01-30T23:50:00+00:00",
		"public_net": {
			"ipv4": {
				"ip": "1.2.3.4",
				"blocked": false,
				"dns_ptr": "server01.test.invalid"
			},
			"ipv6": {
				"ip": "2001:db8::/64",
				"blocked": false,
				"dns_ptr": [
					{
						"ip": "2001:db8::1",
						"dns_ptr": "server01.test.invalid"
					}
				]
			},
			"floating_ips": [ 42 ]
		},
		"private_net": [
			{
				"network": 1,
				"ip": "10.0.0.2",
				"alias_ips": [],
				"mac_address": "aa:bb:cc:dd:ee:ff"
			}
		],
		"server_type": {
			"id": 1,
			"name": "%s",
			"description": "Test",
			"cores": 1,
			"memory": 1,
			"disk": 25,
			"deprecated": true,
			"prices": [
				{
					"location": "hel1",
					"price_hourly": {
						"net": "1.0000000000",
						"gross": "1.1900000000000000"
					},
					"price_monthly": {
						"net": "1.0000000000",
						"gross": "1.1900000000000000"
					}
				}
			],
			"storage_type": "local",
			"cpu_type": "shared"
		},
		"datacenter": {
			"id": 1,
			"name": "%s",
			"description": "Test",
			"location": {
				"id": 2,
				"name": "hel1",
				"description": "Helsinki DC 2",
				"country": "FI",
				"city": "Helsinki",
				"latitude": 60.1698,
				"longitude": 24.9386,
				"network_zone": "eu-central"
			},
			"server_types": {
				"supported": [ 1, 2, 3 ],
				"available": [ 1, 2, 3 ],
				"available_for_migration": [ 1, 2, 3 ]
			}
		},
		"image": %s,
		"iso": {
			"id": 42,
			"name": "netboot",
			"description": "netboot ISO",
			"type": "public",
			"deprecated": "2018-02-28T00:00:00+00:00"
		},
		"rescue_enabled": false,
		"locked": false,
		"backup_window": "22-02",
		"outgoing_traffic": 123456,
		"ingoing_traffic": 123456,
		"included_traffic": 654321,
		"protection": {
			"delete": false,
			"rebuild": false
		},
		"labels": {},
		"volumes": [],
		"load_balancers": []
	},
	"action": {
		"id": 1,
		"command": "create_server",
		"status": "running",
		"progress": 0,
		"started": "2016-01-30T23:50:00+00:00",
		"finished": null,
		"resources": [
			{
				"id": 42,
				"type": "server"
			}
		],
		"error": {
			"code": "action_failed",
			"message": "Action failed"
		}
	},
	"next_actions": [
	],
	"root_password": "test"
}
			`,
			data["name"],
			data["server_type"],
			data["datacenter"],
			jsonImageData)))
		}
	})
}

func SetupSshKeysEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/ssh_keys", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		queryParams := req.URL.Query()

		res.Write([]byte(`
{
	"ssh_keys": [
		`))

		if (queryParams.Get("name") == TestProviderSpecKeyName) {
			res.Write([]byte(`
{
	"id": 42,
	"name": "Simulated ssh key",
	"fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff",
	"public_key": "ssh-rsa invalid",
	"labels": {},
	"created": "2016-01-30T23:50:00+00:00"
}
			`))
		}

		res.Write([]byte(`
	]
}
		`))
	})
}
