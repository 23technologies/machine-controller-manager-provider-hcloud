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
	"fmt"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProviderID", func() {
	Describe("#DecodeServerDataFromProviderID", func() {
		It("should correctly parse and return decoded server information", func() {
			serverData, err := DecodeServerDataFromProviderID(EncodeProviderID(mock.TestProviderSpecZone, 42))

			Expect(err).NotTo(HaveOccurred())
			Expect(serverData.Zone).To(Equal(mock.TestProviderSpecZone))
			Expect(serverData.ID).To(Equal(42))
		})

		It("should fail if an unsupported provider ID scheme is provided", func() {
			_, err := DecodeServerDataFromProviderID("invalid:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no server ID", func() {
			_, err := DecodeServerDataFromProviderID("hcloud:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no zone", func() {
			_, err := DecodeServerDataFromProviderID("hcloud:///1")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains an invalid server ID", func() {
			_, err := DecodeServerDataFromProviderID("hcloud:///test/nan")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DecodeZoneFromProviderID", func() {
		It("should correctly parse and return a zone", func() {
			zone, err := DecodeZoneFromProviderID(EncodeProviderID(mock.TestProviderSpecZone, 42))

			Expect(err).NotTo(HaveOccurred())
			Expect(zone).To(Equal(mock.TestProviderSpecZone))
		})

		It("should fail if an unsupported provider ID scheme is provided", func() {
			_, err := DecodeZoneFromProviderID("invalid:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no server ID", func() {
			_, err := DecodeZoneFromProviderID("hcloud:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no zone", func() {
			_, err := DecodeZoneFromProviderID("hcloud:///1")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains an invalid server ID", func() {
			_, err := DecodeZoneFromProviderID("hcloud:///test/nan")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DecodeServerIDFromProviderID", func() {
		It("should correctly parse and return a server ID", func() {
			serverID, err := DecodeServerIDFromProviderID(EncodeProviderID(mock.TestProviderSpecZone, 42))

			Expect(err).NotTo(HaveOccurred())
			Expect(serverID).To(Equal(42))
		})

		It("should fail if an unsupported provider ID scheme is provided", func() {
			_, err := DecodeServerIDFromProviderID("invalid:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no server ID", func() {
			_, err := DecodeServerIDFromProviderID("hcloud:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no zone", func() {
			_, err := DecodeServerIDFromProviderID("hcloud:///1")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains an invalid server ID", func() {
			_, err := DecodeServerIDFromProviderID("hcloud:///test/nan")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DecodeServerIDAsStringFromProviderID", func() {
		It("should correctly parse and return a server ID", func() {
			serverID, err := DecodeServerIDAsStringFromProviderID(EncodeProviderID(mock.TestProviderSpecZone, 42))

			Expect(err).NotTo(HaveOccurred())
			Expect(serverID).To(Equal("42"))
		})

		It("should fail if an unsupported provider ID scheme is provided", func() {
			_, err := DecodeServerIDAsStringFromProviderID("invalid:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no server ID", func() {
			_, err := DecodeServerIDAsStringFromProviderID("hcloud:///test")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains no zone", func() {
			_, err := DecodeServerIDAsStringFromProviderID("hcloud:///1")

			Expect(err).To(HaveOccurred())
		})
		It("should fail if a provider ID definition contains an invalid server ID", func() {
			_, err := DecodeServerIDAsStringFromProviderID("hcloud:///test/nan")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#EncodeProviderID", func() {
		It("should correctly encode a provider ID", func() {
			providerID := EncodeProviderID(mock.TestProviderSpecZone, 42)
			Expect(providerID).To(Equal(fmt.Sprintf("hcloud:///%s/%d", mock.TestProviderSpecZone, 42)))
		})
	})
})
