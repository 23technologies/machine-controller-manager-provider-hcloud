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

// Package validation - validation is used to validate cloud specific ProviderSpec
package validation

import (
	"fmt"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Validation", func() {
	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"userData": []byte("dummy-user-data"),
		},
	}

	Describe("#ValidateHCloudProviderSpec", func() {
		type setup struct {
		}

		type action struct {
			spec   *apis.ProviderSpec
			secret *corev1.Secret
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
				errList := ValidateHCloudProviderSpec(data.action.spec, data.action.secret)

				if data.expect.errToHaveOccurred {
					Expect(errList).NotTo(BeNil())
					Expect(errList).To(Equal(data.expect.errList))
				} else {
					Expect(errList).To(BeEmpty())
				}
			},

			Entry("Simple validation of HCloud machine class", &data{
				setup: setup{},
				action: action{
					spec: mock.NewProviderSpec(),
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
			Entry("cluster field missing", &data{
				setup: setup{},
				action: action{
					spec: &apis.ProviderSpec{
						Zone: mock.TestProviderSpecDatacenter,
						ImageName: mock.TestProviderSpecImageName,
						ServerType: mock.TestProviderSpecServerType,
						SSHFingerprint: mock.TestProviderSpecSSHFingerprint,
					},
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: true,
					errList: []error{
						fmt.Errorf("cluster is required field"),
					},
				},
			}),
			Entry("zone field missing", &data{
				setup: setup{},
				action: action{
					spec: &apis.ProviderSpec{
						Cluster: mock.TestProviderSpecCluster,
						ImageName: mock.TestProviderSpecImageName,
						ServerType: mock.TestProviderSpecServerType,
						SSHFingerprint: mock.TestProviderSpecSSHFingerprint,
					},
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: true,
					errList: []error{
						fmt.Errorf("zone is required field"),
					},
				},
			}),
			Entry("imageName field missing", &data{
				setup: setup{},
				action: action{
					spec: &apis.ProviderSpec{
						Cluster: mock.TestProviderSpecCluster,
						Zone: mock.TestProviderSpecDatacenter,
						ServerType: mock.TestProviderSpecServerType,
						SSHFingerprint: mock.TestProviderSpecSSHFingerprint,
					},
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: true,
					errList: []error{
						fmt.Errorf("imageName is required field"),
					},
				},
			}),
			Entry("serverType field missing", &data{
				setup: setup{},
				action: action{
					spec: &apis.ProviderSpec{
						Cluster: mock.TestProviderSpecCluster,
						Zone: mock.TestProviderSpecDatacenter,
						ImageName: mock.TestProviderSpecImageName,
						SSHFingerprint: mock.TestProviderSpecSSHFingerprint,
					},
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: true,
					errList: []error{
						fmt.Errorf("serverType is required field"),
					},
				},
			}),
			Entry("sshFingerprint field missing", &data{
				setup: setup{},
				action: action{
					spec: &apis.ProviderSpec{
						Cluster: mock.TestProviderSpecCluster,
						Zone: mock.TestProviderSpecDatacenter,
						ImageName: mock.TestProviderSpecImageName,
						ServerType: mock.TestProviderSpecServerType,
					},
					secret: providerSecret,
				},
				expect: expect{
					errToHaveOccurred: true,
					errList: []error{
						fmt.Errorf("sshFingerprint is required field"),
					},
				},
			}),
		)
	})
})
