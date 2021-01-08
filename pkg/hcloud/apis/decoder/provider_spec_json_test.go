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

// Package decoder is used for API related object transformations
package decoder

import (
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Decoder", func() {
	machineClass := &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: []byte("{\"imageName\":\"ubuntu-20.04\",\"serverType\":\"cx11-ceph\",\"datacenter\":\"hel1-dc2\",\"keyName\":\"test-ssh-publickey\"}"),
		},
	}
	unsupportedMachineClass := &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: []byte("{\"data\":[]}"),
		},
	}

	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"userData": []byte("dummy-user-data"),
		},
	}

	Describe("#DecodeProviderSpecFromMachineClass", func() {
		It("should correctly parse and return a ProviderSpec object", func() {
			providerSpec, err := DecodeProviderSpecFromMachineClass(machineClass, providerSecret)

			Expect(err).NotTo(HaveOccurred())
			Expect(providerSpec.KeyName).To(Equal("test-ssh-publickey"))
		})
		It("should fail of an invalid machineClass is provided", func() {
			_, err := DecodeProviderSpecFromMachineClass(unsupportedMachineClass, providerSecret)

			Expect(err).To(HaveOccurred())
		})
	})
})
