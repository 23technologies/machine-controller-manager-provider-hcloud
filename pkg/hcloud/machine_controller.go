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

	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/machine-controller-manager-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog"
	"strconv"
)

// CreateMachine handles a machine creation request
//
// PARAMETERS
// Machine      *v1alpha1.Machine      Machine object from whom VM is to be created
// MachineClass *v1alpha1.MachineClass MachineClass backing the machine object
// Secret       *corev1.Secret         Kubernetes secret that contains any sensitive data/credentials
//
func (p *MachineProvider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	var (
		machine      = req.Machine
		secret       = req.Secret
		machineClass = req.MachineClass
	)
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", req.Machine.Name)

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userData, ok := secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, "userData doesn't exist")
	}

	userDataBase64Enc := base64.StdEncoding.EncodeToString([]byte(userData))
	token := string(req.Secret.Data["token"])

	client := api.GetClientForToken(token)

	imageName := providerSpec.ImageName
	image, _, err := client.Image.Get(ctx, imageName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if image == nil {
		images, err := client.Image.AllWithOpts(ctx, hcloud.ImageListOpts{Name: imageName, IncludeDeprecated: true})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if len(images) == 0 {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Image %s not found", imageName))
		} else {
			image = images[0]
		}
	}

	name := machine.Name
	serverType := providerSpec.ServerType
	labels := map[string]string{"createdby": "gardener", "test": "test"}
	startAfterCreate := true

	opts := hcloud.ServerCreateOpts{
		Name: name,
		ServerType: &hcloud.ServerType{
			Name: serverType,
		},
		Image:            image,
		Labels:           labels,
		Datacenter:       &hcloud.Datacenter{Name: providerSpec.Datacenter},
		UserData:         userDataBase64Enc,
		StartAfterCreate: &startAfterCreate,
	}

	var sshKey *hcloud.SSHKey
	sshKey, _, err = client.SSHKey.Get(ctx, providerSpec.KeyName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	opts.SSHKeys = append(opts.SSHKeys, sshKey)

	server, _, err := client.Server.Create(ctx, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &driver.CreateMachineResponse{
		ProviderID: strconv.Itoa(server.Server.ID),
		NodeName:   server.Server.Name,
	}

	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (driver.DeleteMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be deleted
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.DeleteMachineResponse)
// LastKnownState        bytes(blob)              (Optional) Last known state of VM during the current operation.
//                                                Could be helpful to continue operations in future requests.
//
func (p *MachineProvider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	token := string(req.Secret.Data["token"])
	client := api.GetClientForToken(token)

	server, _, err := client.Server.Get(ctx, req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if server == nil {
		klog.V(3).Infof("VM %q for Machine %q did not exist", req.Machine.Spec.ProviderID, req.Machine.Name)
		return &driver.DeleteMachineResponse{}, nil
	}

	_, err = client.Server.Delete(ctx, server)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.V(3).Infof("VM %q for Machine %q was terminated succesfully", req.Machine.Spec.ProviderID, req.Machine.Name)
	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.GetMachineStatusRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM status needs to be returned
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.GetMachineStatueResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                                ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                                Eg: gce://project-name/region/vm-ProviderID
// NodeName             string                    Returns the name of the node-object that the VM register's with Kubernetes.
//                                                This could be different from req.MachineName as well
//
// The request should return a NOT_FOUND (5) status error code if the machine is not existing
func (p *MachineProvider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)

	token := string(req.Secret.Data["token"])
	client := api.GetClientForToken(token)

	server, _, err := client.Server.Get(ctx, req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if server == nil {
		klog.V(3).Infof("VM %q for Machine %q did not exist", req.Machine.Spec.ProviderID, req.Machine.Name)
		return &driver.GetMachineStatusResponse{}, status.Error(codes.NotFound, "")
	}

	response := &driver.GetMachineStatusResponse{
		NodeName:   server.Name,
		ProviderID: strconv.Itoa(server.ID),
	}

	klog.V(3).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)
	return response, nil
	// return &driver.GetMachineStatusResponse{}, status.Error(codes.Unimplemented, "")

}

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// MachineClass          *v1alpha1.MachineClass   MachineClass based on which VMs created have to be listed
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//                                           for all machine's who where possibilly created by this ProviderSpec
//
func (p *MachineProvider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been recieved for %q", req.MachineClass.Name)
	defer klog.V(2).Infof("List machines request has been processed for %q", req.MachineClass.Name)

	token := string(req.Secret.Data["token"])
	client := api.GetClientForToken(token)

	listopts := hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: "createdby=gardener",
			PerPage:       50,
		},
	}
	servers, err := client.Server.AllWithOpts(ctx, listopts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	listOfVMs := make(map[string]string)

	for _, server := range servers {
		listOfVMs[strconv.Itoa(server.ID)] = server.Name
	}

	resp := &driver.ListMachinesResponse{
		MachineList: listOfVMs,
	}

	return resp, nil
	// return &driver.ListMachinesResponse{}, status.Error(codes.Unimplemented, "")
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (driver.GetVolumeIDsRequest)
// PVSpecList            []*corev1.PersistentVolumeSpec       PVSpecsList is a list PV specs for whom volume-IDs are required.
//
// RESPONSE PARAMETERS (driver.GetVolumeIDsResponse)
// VolumeIDs             []string                             VolumeIDs is a repeated list of VolumeIDs.
//
func (p *MachineProvider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been recieved for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "")
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
// For instance an machineClass custom resource of `AWSMachineClass` to `MachineClass`.
// Implement this functionality only if something like this is desired in your setup.
// If you don't require this functionality leave is as is. (return Unimplemented)
//
// The following are the tasks typically expected out of this method
// 1. Validate if the incoming classSpec is valid one for migration (e.g. has the right kind).
// 2. Migrate/Copy over all the fields/spec from req.ProviderSpecificMachineClass to req.MachineClass
// For an example refer
//		https://github.com/prashanth26/machine-controller-manager-provider-gcp/blob/migration/pkg/gcp/machine_controller.go#L222-L233
//
// REQUEST PARAMETERS (driver.GenerateMachineClassForMigration)
// ProviderSpecificMachineClass    interface{}                             ProviderSpecificMachineClass is provider specfic machine class object (E.g. AWSMachineClass). Typecasting is required here.
// MachineClass 				   *v1alpha1.MachineClass                  MachineClass is the machine class object that is to be filled up by this method.
// ClassSpec                       *v1alpha1.ClassSpec                     Somemore classSpec details useful while migration.
//
// RESPONSE PARAMETERS (driver.GenerateMachineClassForMigration)
// NONE
//
func (p *MachineProvider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been recieved for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}
