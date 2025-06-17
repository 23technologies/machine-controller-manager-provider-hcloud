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
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin", func() {
	Describe("#NewHCloudProvider", func() {
		It("should correctly create a new provider object", func() {
			provider := NewHCloudProvider()
			Expect(provider).NotTo(BeNil(), "NewHCloudProvider should not return nil")
			_, ok := provider.(driver.Driver)
			Expect(ok).To(BeTrue(), "The returned provider should implement driver.Driver")
		})
	})
})
