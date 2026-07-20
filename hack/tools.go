// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package tools

import (
	_ "github.com/gardener/gardener/.github"
	_ "github.com/gardener/gardener/.github/ISSUE_TEMPLATE"
	_ "github.com/gardener/gardener/hack/.ci"
	_ "github.com/gardener/gardener/hack/tools"

	_ "github.com/ahmetb/gen-crd-api-reference-docs"
	_ "go.uber.org/mock/mockgen"
	_ "k8s.io/code-generator"
)
