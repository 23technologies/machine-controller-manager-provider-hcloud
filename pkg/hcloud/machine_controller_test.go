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

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/opendefensecloud/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	"github.com/opendefensecloud/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/mock"
)

var provider *MachineProvider

var _ = BeforeSuite(func() {
	provider = &MachineProvider{}
})

var _ = Describe("MachineController", func() {
	var mockTestEnv mock.MockTestEnv
	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"token":    []byte("dummy-token"),
			"userData": []byte("dummy-user-data"),
		},
	}

	var _ = BeforeEach(func() {
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.Client)
		mock.SetupFloatingIPsEndpointOnMux(mockTestEnv.Mux)
		mock.SetupImagesEndpointOnMux(mockTestEnv.Mux)
		mock.SetupServersEndpointOnMux(mockTestEnv.Mux, true)
		mock.SetupSshKeysEndpointOnMux(mockTestEnv.Mux)
		mock.SetupTestPlacementGroupEndpointOnMux(mockTestEnv.Mux)
		mock.SetupTestServerEndpointOnMux(mockTestEnv.Mux)
		mock.ResetPlacementGroupHitCounters()
	})

	var _ = AfterEach(func() {
		mockTestEnv.Teardown()
		apis.SetClientForToken("dummy-token", nil)
	})

	Describe("#CreateMachine", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.CreateMachineRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errStatus         codes.Code
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

					errStatus, ok := err.(*status.Status)
					Expect(ok).To(BeTrue())
					Expect(errStatus.Code()).To(Equal(data.expect.errStatus))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("is correctly executed", &data{
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

			Entry("contains a provider ID", &data{
				setup: setup{},
				action: action{
					&driver.CreateMachineRequest{
						Machine:      mock.NewMachine(mock.TestServerID),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
				},
			}),
			Entry("contains an invalid provider ID", &data{
				setup: setup{},
				action: action{
					&driver.CreateMachineRequest{
						Machine:      mock.ManipulateMachine(mock.NewMachine(mock.TestServerID), map[string]interface{}{"Spec.ProviderID": "test:///invalid"}),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
				},
			}),
			Entry("contains an invalid machine class", &data{
				setup: setup{},
				action: action{
					&driver.CreateMachineRequest{
						Machine:      mock.NewMachine(-1),
						MachineClass: mock.NewMachineClassWithProviderSpec([]byte(mock.TestInvalidProviderSpec)),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
				},
			}),
		)

		It("applies the placement group when PlacementGroupID is set without a floating pool", func() {
			ctx := context.Background()

			// TestProviderSpec carries placementGroupID "42" and no floatingPoolName,
			// so the placement-group flow must be driven by PlacementGroupID alone.
			_, err := provider.CreateMachine(ctx, &driver.CreateMachineRequest{
				Machine:      mock.NewMachine(-1),
				MachineClass: mock.NewMachineClass(),
				Secret:       providerSecret,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(mock.PlacementGroupGetByIDHitCount()).To(BeNumerically(">", 0), "expected GET /placement_groups/42 to be hit")
			Expect(mock.AddToPlacementGroupHitCount()).To(BeNumerically(">", 0), "expected add_to_placement_group action to be hit")
		})

		It("does not delete a foreign server that merely shares a name (never deletes by name)", func() {
			ctx := context.Background()

			env := mock.NewMockTestEnv()
			defer env.Teardown()

			const token = "cleanup-foreign-token"
			apis.SetClientForToken(token, env.Client)
			defer apis.SetClientForToken(token, nil)

			mock.SetupImagesEndpointOnMux(env.Mux)
			mock.SetupSshKeysEndpointOnMux(env.Mux)
			mock.SetupFloatingIPsEndpointOnMux(env.Mux)
			// emptyOnFirstRequest=false: GetByName finds a server named "machine-0"
			// immediately. That fixture server carries no labels, so it does not match
			// this request's providerSpec labels and createMachineValidateDeploymentAndCleanup
			// correctly treats it as foreign (not created by this request).
			mock.SetupServersEndpointOnMux(env.Mux, false)
			mock.SetupTestServerEndpointOnMux(env.Mux)
			mock.ResetServerDeleteHitCounter()

			secret := &corev1.Secret{
				Data: map[string][]byte{
					"token":    []byte(token),
					"userData": []byte("dummy-user-data"),
				},
			}

			_, err := provider.CreateMachine(ctx, &driver.CreateMachineRequest{
				Machine:      mock.NewMachine(-1),
				MachineClass: mock.NewMachineClass(),
				Secret:       secret,
			})

			Expect(err).To(HaveOccurred())
			Expect(mock.ServerDeleteHitCount()).To(Equal(int32(0)), "create-error cleanup must never delete a server found only by name")
		})
	})

	Describe("#DeleteMachine", func() {
		type setup struct {
		}

		type action struct {
			machineRequest *driver.DeleteMachineRequest
		}

		type expect struct {
			errToHaveOccurred bool
			errStatus         codes.Code
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

					errStatus, ok := err.(*status.Status)
					Expect(ok).To(BeTrue())
					Expect(errStatus.Code()).To(Equal(data.expect.errStatus))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("is correctly executed", &data{
				setup: setup{},
				action: action{
					&driver.DeleteMachineRequest{
						Machine:      mock.NewMachine(mock.TestServerID),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),

			Entry("contains no provider ID", &data{
				setup: setup{},
				action: action{
					&driver.DeleteMachineRequest{
						Machine:      mock.NewMachine(-1),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
			Entry("contains an invalid provider ID", &data{
				setup: setup{},
				action: action{
					&driver.DeleteMachineRequest{
						Machine:      mock.ManipulateMachine(mock.NewMachine(mock.TestServerID), map[string]interface{}{"Spec.ProviderID": "test:///invalid"}),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
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
			errStatus         codes.Code
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

					errStatus, ok := err.(*status.Status)
					Expect(ok).To(BeTrue())
					Expect(errStatus.Code()).To(Equal(data.expect.errStatus))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("is correctly executed", &data{
				setup: setup{},
				action: action{
					&driver.GetMachineStatusRequest{
						Machine:      mock.NewMachine(mock.TestServerID),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),

			Entry("contains no provider ID", &data{
				setup: setup{},
				action: action{
					&driver.GetMachineStatusRequest{
						Machine:      mock.NewMachine(-1),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.NotFound,
				},
			}),
			Entry("contains an invalid provider ID", &data{
				setup: setup{},
				action: action{
					&driver.GetMachineStatusRequest{
						Machine:      mock.ManipulateMachine(mock.NewMachine(mock.TestServerID), map[string]interface{}{"Spec.ProviderID": "test:///invalid"}),
						MachineClass: mock.NewMachineClass(),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
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
			errStatus         codes.Code
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

					errStatus, ok := err.(*status.Status)
					Expect(ok).To(BeTrue())
					Expect(errStatus.Code()).To(Equal(data.expect.errStatus))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("is correctly executed", &data{
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
			Entry("contains an invalid machine class", &data{
				setup: setup{},
				action: action{
					&driver.ListMachinesRequest{
						MachineClass: mock.NewMachineClassWithProviderSpec([]byte(mock.TestInvalidProviderSpec)),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errStatus:         codes.InvalidArgument,
				},
			}),
		)
	})

	Describe("#GetVolumeIDs", func() {
		It("is not implemented", func() {
			ctx := context.Background()

			req := &driver.GetVolumeIDsRequest{
				PVSpecs: []*corev1.PersistentVolumeSpec{
					{
						Capacity:                      map[corev1.ResourceName]resource.Quantity{},
						PersistentVolumeSource:        corev1.PersistentVolumeSource{},
						AccessModes:                   []corev1.PersistentVolumeAccessMode{},
						ClaimRef:                      &corev1.ObjectReference{},
						PersistentVolumeReclaimPolicy: "",
						StorageClassName:              "",
						MountOptions:                  []string{},
						NodeAffinity:                  &corev1.VolumeNodeAffinity{},
					},
				},
			}

			_, err := provider.GetVolumeIDs(ctx, req)
			Expect(err).To(HaveOccurred())

			errStatus, ok := err.(*status.Status)
			Expect(ok).To(BeTrue())
			Expect(errStatus.Code()).To(Equal(codes.Unimplemented))
		})
	})
})
