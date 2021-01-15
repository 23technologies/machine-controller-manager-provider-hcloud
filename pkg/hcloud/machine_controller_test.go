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

// Package hcloud contains the HCloud provider specific implementations to manage machines
package hcloud

import (
	"context"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/mock"
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MachineController", func() {
	var mockTestEnv mock.MockTestEnv
	var provider *MachineProvider
	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"token":    []byte("dummy-token"),
			"userData": []byte("dummy-user-data"),
		},
	}

	var _ = BeforeSuite(func() {
		provider = &MachineProvider {
			SPI: &spi.PluginSPIImpl{},
		}
	})

	var _ = BeforeEach(func() {
		mockTestEnv = mock.NewMockTestEnv()

		api.SetClientForToken("dummy-token", mockTestEnv.Client)
		mock.SetupImagesEndpointOnMux(mockTestEnv.Mux)
		mock.SetupServersEndpointOnMux(mockTestEnv.Mux)
		mock.SetupServer42EndpointOnMux(mockTestEnv.Mux)
		mock.SetupSshKeysEndpointOnMux(mockTestEnv.Mux)
	})

	var _ = AfterEach(func() {
		mockTestEnv.Teardown()
		api.SetClientForToken("dummy-token", nil)
	})

	Describe("#CreateMachine", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.CreateMachineRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errList           []error
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.Background()
				_, err := provider.CreateMachine(ctx, data.action.machineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.errList))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("Valid use case", &data{
				setup: setup{},
				action: action{
					&driver.CreateMachineRequest{
						Machine:      mock.NewMachine(-1),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
		)
	})

	Describe("#DeleteMachine", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.DeleteMachineRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errList           []error
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.Background()
				_, err := provider.DeleteMachine(ctx, data.action.machineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.errList))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("Valid use case", &data{
				setup: setup{},
				action: action{
					&driver.DeleteMachineRequest{
						Machine:      mock.NewMachine(42),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
		)
	})

	Describe("#GetMachineStatus", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.GetMachineStatusRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errList           []error
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.Background()
				_, err := provider.GetMachineStatus(ctx, data.action.machineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.errList))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("Valid use case", &data{
				setup: setup{},
				action: action{
					&driver.GetMachineStatusRequest{
						Machine:      mock.NewMachine(42),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
		)
	})

	Describe("#ListMachines", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.ListMachinesRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errList           []error
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.Background()
				_, err := provider.ListMachines(ctx, data.action.machineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.errList))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("Valid use case", &data{
				setup: setup{},
				action: action{
					&driver.ListMachinesRequest{
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
		)
	})
})
