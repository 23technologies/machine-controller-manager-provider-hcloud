/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.

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
	api "github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

// ValidateProviderSpecNSecret validates provider spec and secret to check if all fields are present and valid
func ValidateProviderSpecNSecret(spec *api.ProviderSpec, secrets *corev1.Secret) []error {
	var allErrs []error

	if "" == spec.ImageName {
		allErrs = append(allErrs, fmt.Errorf("imageName is required field"))
	}
	if "" == spec.ServerType {
		allErrs = append(allErrs, fmt.Errorf("serverType is required field"))
	}
	if "" == spec.Datacenter {
		allErrs = append(allErrs, fmt.Errorf("datacenter is required field"))
	}
	if "" == spec.KeyName {
		allErrs = append(allErrs, fmt.Errorf("keyName is required field"))
	}
	//allErrs = append(allErrs, ValidateSecret(secret)...)

	return allErrs
}
