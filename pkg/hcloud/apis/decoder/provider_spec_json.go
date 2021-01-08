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
	"encoding/json"
	"fmt"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	validation "github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/validation"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
)

// DecodeProviderSpecFromMachineClass decodes the given MachineClass to receive the ProviderSpec.
//
// PARAMETERS
// machineClass *v1alpha1.MachineClass MachineClass backing the machine object
// secret       *corev1.Secret         Kubernetes secret that contains any sensitive data/credentials
func DecodeProviderSpecFromMachineClass(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (*api.ProviderSpec, error) {
	var (
		providerSpec *api.ProviderSpec
	)

	// Extract providerSpec
	if machineClass == nil {
		return nil, status.Error(codes.Internal, "MachineClass ProviderSpec is nil")
	}

	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Validate the Spec
	ValidationErr := validation.ValidateHCloudProviderSpec(providerSpec, secret)
	if ValidationErr != nil {
		err = fmt.Errorf("Error while validating ProviderSpec %v", ValidationErr)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}
