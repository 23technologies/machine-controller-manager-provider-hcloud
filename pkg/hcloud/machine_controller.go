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
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog"
)

// Constant defaultMachineOperationInterval is the time to wait between retries
const defaultMachineOperationInterval = 5 * time.Second

// CreateMachine handles a machine creation request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The create request for VM creation
func (p *MachineProvider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	var (
		machine      = req.Machine
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", machine.Name)

	if "" != machine.Spec.ProviderID {
		return nil, status.Error(codes.InvalidArgument, "Machine creation with existing provider ID is not supported")
	}

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userData, ok := secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, "userData doesn't exist")
	}

	client := apis.GetClientForToken(string(secret.Data["token"]))

	imageName := providerSpec.ImageName
	userDataBase64Enc := base64.StdEncoding.EncodeToString([]byte(userData))

	image, _, err := client.Image.GetByName(ctx, imageName)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if image == nil {
		images, err := client.Image.AllWithOpts(ctx, hcloud.ImageListOpts{Name: imageName, IncludeDeprecated: true})
		if nil != err {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if len(images) == 0 {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Image %s not found", imageName))
		} else {
			image = images[0]
		}
	}

	region := apis.GetRegionFromZone(providerSpec.Zone)
	startAfterCreate := false
	zone := providerSpec.Zone

	opts := hcloud.ServerCreateOpts{
		Name:             machine.Name,
		ServerType:       &hcloud.ServerType{Name: providerSpec.ServerType},
		Image:            image,
		Labels:           map[string]string{
			"mcm.gardener.cloud/cluster": providerSpec.Cluster,
			"mcm.gardener.cloud/role": "node",
			"topology.kubernetes.io/region": region,
			"topology.kubernetes.io/zone": zone,
		},
		Datacenter:       &hcloud.Datacenter{Name: zone},
		UserData:         userDataBase64Enc,
		StartAfterCreate: &startAfterCreate,
	}

	sshKey, _, err := client.SSHKey.GetByFingerprint(ctx, providerSpec.SSHFingerprint)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	} else if sshKey == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("SSH key with fingerprint %s not found", providerSpec.SSHFingerprint))
	}

	opts.SSHKeys = append(opts.SSHKeys, sshKey)

	if "" != providerSpec.NetworkName {
		network, _, err := client.Network.GetByName(ctx, providerSpec.NetworkName)
		if nil != err {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else if network == nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Network %s not found", providerSpec.NetworkName))
		}

		opts.Networks = append(opts.Networks, network)
	}

	serverResult, _, err := client.Server.Create(ctx, opts)
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	server := serverResult.Server
	repeat := server.Locked || hcloud.ServerStatusInitializing == server.Status

	for repeat {
		time.Sleep(defaultMachineOperationInterval)

		server, _, err = client.Server.GetByName(ctx, machine.Name)
		if nil != err {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		repeat = server.Locked || hcloud.ServerStatusInitializing == server.Status
	}

	if "" != providerSpec.FloatingPoolName {
		name := fmt.Sprintf("%s-%s-ipv4", providerSpec.FloatingPoolName, machine.Name)

		floatingIP, _, err := client.FloatingIP.GetByName(ctx, name)
		if nil != err {
			return nil, status.Error(codes.Internal, err.Error())
		} else if floatingIP == nil {
			opts := hcloud.FloatingIPCreateOpts{
				Name: &name,
				Type: hcloud.FloatingIPTypeIPv4,
				Server: server,
				Labels: map[string]string{
					"mcm.gardener.cloud/cluster": providerSpec.Cluster,
					"networking.hcloud.mcm.gardener.cloud/floating-pool": providerSpec.FloatingPoolName,
				},
			}

			_, _, err := client.FloatingIP.Create(ctx, opts)
			if nil != err {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}

	if hcloud.ServerStatusStarting != server.Status && hcloud.ServerStatusRunning != server.Status {
		_, _, err = client.Server.Poweron(ctx, server)
		if nil != err {
			return nil, status.Error(codes.Aborted, err.Error())
		}
	}

	response := &driver.CreateMachineResponse{
		ProviderID: transcoder.EncodeProviderID(providerSpec.Zone, server.ID),
		NodeName:   server.Name,
	}

	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The delete request for VM deletion
func (p *MachineProvider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	var (
		machine      = req.Machine
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", machine.Name)

	serverID, err := transcoder.DecodeServerIDFromProviderID(machine.Spec.ProviderID)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := apis.GetClientForToken(string(secret.Data["token"]))

	server, _, err := client.Server.GetByID(ctx, serverID)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	} else if server == nil {
		klog.V(3).Infof("VM %s (%d) does not exist", machine.Name, serverID)
		return &driver.DeleteMachineResponse{}, nil
	}

	_, err = client.Server.Delete(ctx, server)
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	name := fmt.Sprintf("%s-%s-ipv4", providerSpec.FloatingPoolName, machine.Name)

	floatingIP, _, err := client.FloatingIP.GetByName(ctx, name)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	} else if nil != floatingIP {
		_, err = client.FloatingIP.Delete(ctx, floatingIP)
		if nil != err {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}

	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The get request for VM info
func (p *MachineProvider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	var (
		machine      = req.Machine
		secret       = req.Secret
	)

	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", machine.Name)

	// Handle case where machine lookup occurs with empty provider ID
	if machine.Spec.ProviderID == "" {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Provider ID for machine %q is not defined", machine.Name))
	}

	serverID, err := transcoder.DecodeServerIDFromProviderID(machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := apis.GetClientForToken(string(secret.Data["token"]))

	server, _, err := client.Server.GetByID(ctx, serverID)
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	} else if server == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("VM %s (%d) does not exist", machine.Name, serverID))
	}

	return &driver.GetMachineStatusResponse{ ProviderID: machine.Spec.ProviderID, NodeName: server.Name }, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request object to get a list of VMs belonging to a machineClass
func (p *MachineProvider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	var (
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been received for %q", machineClass.Name)
	defer klog.V(2).Infof("List machines request has been processed for %q", machineClass.Name)

	client := apis.GetClientForToken(string(secret.Data["token"]))
	zone := providerSpec.Zone

	listopts := hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: fmt.Sprintf(
				"mcm.gardener.cloud/cluster=%s,mcm.gardener.cloud/role=node,topology.kubernetes.io/zone=%s",
				url.QueryEscape(providerSpec.Cluster),
				url.QueryEscape(zone),
			),
			PerPage:       50,
		},
	}

	servers, err := client.Server.AllWithOpts(ctx, listopts)
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	listOfVMs := make(map[string]string)

	for _, server := range servers {
		listOfVMs[transcoder.EncodeProviderID(zone, server.ID)] = server.Name
	}

	return &driver.ListMachinesResponse{ MachineList: listOfVMs }, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request object to get a list of VolumeIDs for a PVSpec
func (p *MachineProvider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "")
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request for generating the generic machineClass
func (p *MachineProvider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been received for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}
