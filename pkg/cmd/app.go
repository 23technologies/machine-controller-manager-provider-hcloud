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

// Package cmd provides the provider manager
package cmd

import (
	_ "github.com/gardener/machine-controller-manager/pkg/util/client/metrics/prometheus" // for client metric registration
	"github.com/gardener/machine-controller-manager/pkg/util/provider/app"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/app/options"
	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version/verflag"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud"
)

// RunProviderHCloudManager runs the HCloud machine controller server.
//
// PARAMETERS
// args *pflag.FlagSet Command line arguments
func RunProviderHCloudManager(args *pflag.FlagSet) error {
	s := options.NewMCServer()

	s.AddFlags(args)
	flag.InitFlags()

	verflag.PrintAndExitIfRequested()

	logs.InitLogs()
	defer logs.FlushLogs()

	return app.Run(s, hcloud.NewHCloudProvider())
}
