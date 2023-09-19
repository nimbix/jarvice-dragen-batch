/*
Copyright (c) 2023, Nimbix, Inc.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation are
those of the authors and should not be interpreted as representing official
policies, either expressed or implied, of Nimbix, Inc.
*/

package google

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"jarvice.io/dragen/config"
	"jarvice.io/dragen/internal/logger"
)

func createInstanceClient() (context.Context, *compute.InstancesClient, error) {
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ctx, instancesClient, nil
}

func createSubnetClient() (context.Context, *compute.SubnetworksClient, error) {
	ctx := context.Background()
	subnetClient, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ctx, subnetClient, nil
}

func createReservationsClient() (context.Context, *compute.ReservationsClient, error) {
	ctx := context.Background()
	reservationClient, err := compute.NewReservationsRESTClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ctx, reservationClient, nil
}

func createTemplatesClient() (context.Context, *compute.InstanceTemplatesClient, error) {

	ctx := context.Background()
	templateClient, err := compute.NewInstanceTemplatesRESTClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ctx, templateClient, nil
}

type GoogleCompute struct {
	project, zone, network, name, id string
}

func NewGoogleCompute() (*GoogleCompute, error) {
	var project, zone, network, name, id string
	var err error
	if project, err = googleMetadata("/project/project-id"); err != nil {
		return nil, err
	}
	if zone, err = googleMetadata("/instance/zone"); err != nil {
		return nil, err
	} else {
		zone = filepath.Base(zone)
	}
	if network, err = googleMetadata("/instance/network-interfaces/0/network"); err != nil {
		return nil, err
	} else {
		network = filepath.Base(network)
	}
	if name, err = googleMetadata("/instance/name"); err != nil {
		return nil, err
	}
	if id, err = googleMetadata("/instance/id"); err != nil {
		return nil, err
	}
	return &GoogleCompute{
		project: project,
		zone:    zone,
		network: network,
		name:    name,
		id:      id,
	}, nil
}

func (vm GoogleCompute) GetName() string {
	return vm.name
}

func (vm GoogleCompute) GetProject() string {
	return vm.project
}

func (vm GoogleCompute) GetZone() string {
	return vm.zone
}

func (vm GoogleCompute) GetId() string {
	return vm.id
}

func (vm GoogleCompute) CreateInstanceContainer(name, template, container, shutdownScript, containerCmd string, containerArgs ...string) error {

	args := []string{
		"compute", "instances", "create-with-container", name,
		"--project", vm.project,
		"--zone", vm.zone,
		"--source-instance-template", template,
		"--metadata shutdown-script=\"" + shutdownScript + "\"",
		"--container-image", container,
		"--container-restart-policy", "always",
		"--container-command", containerCmd,
		"--no-shielded-secure-boot",
		"--shielded-vtpm",
		"--shielded-integrity-monitoring",
		"--labels",
		fmt.Sprintf("container-vm=%s", config.DragenImage),
	}
	for _, arg := range containerArgs {
		args = append(args, fmt.Sprintf("--container-arg=%s", arg))
	}
	cmd := exec.Command("/usr/bin/gcloud", args...)

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	err := cmd.Run()
	if err != nil {
		return err
	}

	logger.Ologger.Debug(stdBuffer.String())

	logger.Ologger.Info(name + " container VM created")

	return nil
}

func (vm GoogleCompute) CreateInstance(name, template, reservation, shutdownScript string) error {
	return vm.CreateInstanceWithJobId(name, template, reservation, "", shutdownScript)
}

func (vm GoogleCompute) CreateInstanceWithJobId(name, template, reservation, jobid, shutdownScript string) error {

	ctx, instanceClient, err := createInstanceClient()
	if err != nil {
		return err
	}
	defer instanceClient.Close()

	myTemplate := "/projects/" + vm.project + "/global/instanceTemplates/" + template
	consumeReservationType := "SPECIFIC_RESERVATION"
	key := "compute.googleapis.com/reservation-name"
	shutdown := "shutdown-script"

	metadata := []*computepb.Items{}
	if templateMeta, err := vm.GetTemplateInstance(name); err != nil {
		return err
	} else {
		metadata = templateMeta.Properties.Metadata.Items
	}

	items := []*computepb.Items{
		&computepb.Items{
			Key:   &shutdown,
			Value: &shutdownScript,
		},
	}
	for _, item := range metadata {
		if len(jobid) > 0 {
			if *item.Key == "gce-container-declaration" {
				re := regexp.MustCompile(`(TEMP_JOB_ID)`)
				s := re.ReplaceAllString(*item.Value, jobid)
				item.Value = &s
			}
		}
		items = append(items, item)
	}

	req := &computepb.InsertInstanceRequest{
		InstanceResource: &computepb.Instance{
			Name: &name,
			ReservationAffinity: &computepb.ReservationAffinity{
				ConsumeReservationType: &consumeReservationType,
				Key:                    &key,
				Values: []string{
					reservation,
				},
			},
			Metadata: &computepb.Metadata{
				Items: items,
			},
		},
		Project:                vm.project,
		SourceInstanceTemplate: &myTemplate,
		Zone:                   vm.zone,
	}

	op, err := instanceClient.Insert(ctx, req)
	if err != nil {
		return err
	}

	if err = op.Wait(ctx); err != nil {
		return err
	}

	logger.Ologger.Info(name + " instance created")

	return nil
}

func (vm GoogleCompute) DeleteInstance(name string) error {
	return vm.DeleteInstanceWait(name, true)
}

func (vm GoogleCompute) DeleteInstanceWait(name string, wait bool) error {

	ctx, instancesClient, err := createInstanceClient()
	if err != nil {
		return err
	}
	defer instancesClient.Close()

	req := &computepb.DeleteInstanceRequest{
		Instance: name,
		Project:  vm.project,
		Zone:     vm.zone,
	}

	op, err := instancesClient.Delete(ctx, req)
	if err != nil {
		return err
	}

	if wait {
		if err = op.Wait(ctx); err != nil {
			return err
		}
		logger.Ologger.Info(name + " instance deleted")
	}

	return nil
}

func (vm GoogleCompute) DeleteHost() error {

	ctx, instancesClient, err := createInstanceClient()
	if err != nil {
		return err
	}
	defer instancesClient.Close()

	req := &computepb.DeleteInstanceRequest{
		Instance: vm.name,
		Project:  vm.project,
		Zone:     vm.zone,
	}

	op, err := instancesClient.Delete(ctx, req)
	if err != nil {
		return err
	}

	if err = op.Wait(ctx); err != nil {
		return err
	}

	logger.Ologger.Info(vm.name + " instance deleted")

	return nil
}

func (vm GoogleCompute) ListInstances(filter string) error {

	ctx, instancesClient, err := createInstanceClient()
	if err != nil {
		return err
	}
	defer instancesClient.Close()

	name := fmt.Sprint("name =", filter)

	req := &computepb.ListInstancesRequest{
		Project: vm.project,
		Zone:    vm.zone,
		Filter:  &name,
	}

	instances := instancesClient.List(ctx, req)

	for instance, err := instances.Next(); err == nil; instance, err = instances.Next() {
		logger.Ologger.Info("Found instance " + *instance.Name)
	}

	return nil
}

func (vm GoogleCompute) InstanceExist(filter string) bool {
	if ret, err := vm.InstanceExistWithError(filter); err != nil {
		logger.Ologger.Warn(err.Error())
		return false
	} else {
		return ret
	}
}

func (vm GoogleCompute) InstanceExistWithError(filter string) (bool, error) {

	ctx, instancesClient, err := createInstanceClient()
	if err != nil {
		return false, err
	}
	defer instancesClient.Close()

	name := fmt.Sprint("name =", filter)

	req := &computepb.ListInstancesRequest{
		Project: vm.project,
		Zone:    vm.zone,
		Filter:  &name,
	}

	instances := instancesClient.List(ctx, req)

	if _, err := instances.Next(); err == nil {
		return true, nil
	}

	return false, err
}

func (vm GoogleCompute) CreateReservation(name, template string) error {

	ctx, reservationClient, err := createReservationsClient()
	if err != nil {
		return err
	}
	defer reservationClient.Close()

	description := "gcloud golang reservation for dragen"
	shareType := "LOCAL"
	shareSettings := &computepb.ShareSettings{
		ShareType: &shareType,
	}

	count := int64(1)
	myTemplate := "/projects/" + vm.project + "/global/instanceTemplates/" + template
	specificReservation := &computepb.AllocationSpecificSKUReservation{
		Count:                  &count,
		SourceInstanceTemplate: &myTemplate,
	}
	vmTrue := true
	specificReservationRequired := &vmTrue

	reservation := &computepb.Reservation{
		Description:                 &description,
		Name:                        &name,
		ShareSettings:               shareSettings,
		SpecificReservation:         specificReservation,
		SpecificReservationRequired: specificReservationRequired,
		Zone:                        &vm.zone,
	}

	req := &computepb.InsertReservationRequest{
		Project:             vm.project,
		Zone:                vm.zone,
		ReservationResource: reservation,
	}

	op, err := reservationClient.Insert(ctx, req)
	if err != nil {
		return err
	}

	if err = op.Wait(ctx); err != nil {
		return err
	}

	logger.Ologger.Info(name + " reservation created")

	return nil
}

func (vm GoogleCompute) DeleteReservation(name string) error {
	return vm.DeleteReservationWait(name, true)
}

func (vm GoogleCompute) DeleteReservationWait(name string, wait bool) error {

	ctx, reservationClient, err := createReservationsClient()
	if err != nil {
		return err
	}
	defer reservationClient.Close()

	req := &computepb.DeleteReservationRequest{
		Reservation: name,
		Project:     vm.project,
		Zone:        vm.zone,
	}

	op, err := reservationClient.Delete(ctx, req)
	if err != nil {
		return err
	}

	if wait {
		if err = op.Wait(ctx); err != nil {
			return err
		}
		logger.Ologger.Info(name + " reservation deleted")
	}

	return nil
}

func (vm GoogleCompute) getSubnet(network, region string) (string, error) {

	ctx, subnetClient, err := createSubnetClient()
	if err != nil {
		return "", err
	}
	defer subnetClient.Close()

	name := fmt.Sprint("network = ", "\"https://www.googleapis.com/compute/v1/projects/"+vm.project+"/global/networks/"+network+"\"")

	req := &computepb.ListSubnetworksRequest{
		Project: vm.project,
		Region:  region,
		Filter:  &name,
	}

	subnets := subnetClient.List(ctx, req)

	subnet, err := subnets.Next()
	if err != nil {
		return "", err
	}

	return *subnet.Name, nil
}

func (vm GoogleCompute) CreateTemplate(name, serviceAccount string) error {

	ctx, templateClient, err := createTemplatesClient()
	if err != nil {
		return err
	}
	defer templateClient.Close()

	region := strings.Join(strings.Split(vm.zone, "-")[:2], "-")
	subnet, err := vm.getSubnet(vm.network, region)
	if err != nil {
		return err
	}

	description := "gcloud golang template for dragen"
	vmTrue := true
	vmFalse := false
	machine := config.GoogleMachine
	index := int32(0)
	deviceName := "persistent-disk-0"
	mode := "READ_WRITE"
	diskType := "PERSISTENT"
	source := config.DragenDisk
	networkName := "nic0"
	network := "/projects/" + vm.project + "/global/networks/" + vm.network
	subnetwork := "/projects/" + vm.project + "/regions/" + region + "/subnetworks/" + subnet
	accessName := "external-nat"
	networkTier := "PREMIUM"
	networkType := "ONE_TO_ONE_NAT"
	schedulingMaint := "MIGRATE"
	schedulingProv := "STANDARD"
	saEmail := serviceAccount
	properties := &computepb.InstanceProperties{
		CanIpForward: &vmFalse,
		Disks: []*computepb.AttachedDisk{
			&computepb.AttachedDisk{
				AutoDelete: &vmTrue,
				Boot:       &vmTrue,
				DeviceName: &deviceName,
				Index:      &index,
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					SourceImage: &source,
				},
				Mode: &mode,
				Type: &diskType,
			},
		},
		MachineType: &machine,
		NetworkInterfaces: []*computepb.NetworkInterface{
			&computepb.NetworkInterface{
				AccessConfigs: []*computepb.AccessConfig{
					&computepb.AccessConfig{
						Name:        &accessName,
						NetworkTier: &networkTier,
						Type:        &networkType,
					},
				},
				Name:       &networkName,
				Network:    &network,
				Subnetwork: &subnetwork,
			},
		},
		Scheduling: &computepb.Scheduling{
			AutomaticRestart:  &vmTrue,
			OnHostMaintenance: &schedulingMaint,
			Preemptible:       &vmFalse,
			ProvisioningModel: &schedulingProv,
		},
		ServiceAccounts: []*computepb.ServiceAccount{
			&computepb.ServiceAccount{
				Email: &saEmail,
				Scopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
				},
			},
		},
	}

	template := &computepb.InstanceTemplate{
		Description: &description,
		Name:        &name,
		Properties:  properties,
	}

	req := &computepb.InsertInstanceTemplateRequest{
		Project:                  vm.project,
		InstanceTemplateResource: template,
	}

	op, err := templateClient.Insert(ctx, req)
	if err != nil {
		return err
	}

	if err = op.Wait(ctx); err != nil {
		return err
	}

	logger.Ologger.Info(name + " template created")

	return nil
}

func (vm GoogleCompute) CreateInstanceTemplates(name, serviceAccount, container, containerCmd string, containerArgs ...string) error {

	region := strings.Join(strings.Split(vm.zone, "-")[:2], "-")
	subnet, err := vm.getSubnet(vm.network, region)
	if err != nil {
		return err
	}
	templateArgs := [...]string{
		"compute", "instance-templates", "create-with-container", name,
		"--project", vm.project,
		"--machine-type", config.GoogleMachine,
		"--network", vm.network,
		"--subnet", subnet,
		"--region", region,
		"--maintenance-policy", "MIGRATE",
		"--service-account", serviceAccount,
		"--scopes", "https://www.googleapis.com/auth/cloud-platform",
		"--image-project", config.DragenProject,
		"--image", config.DragenImage,
		"--create-disk", "auto-delete=yes,device-name=dargen-1,mode=rw,size=10,type=pd-balanced",
		"--no-shielded-secure-boot",
		"--shielded-vtpm",
		"--shielded-integrity-monitoring",
		"--labels", fmt.Sprint("container-vm=", config.DragenImage),
		"--container-image", container,
		"--container-restart-policy", "always",
		"--container-command", containerCmd,
	}

	var args []string
	for _, arg := range templateArgs {
		args = append(args, arg)
	}
	for _, arg := range containerArgs {
		args = append(args, fmt.Sprintf("--container-arg=%s", arg))
	}

	cmd := exec.Command("/usr/bin/gcloud", args...)

	var stdBuffer bytes.Buffer
	mw := io.Writer(&stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Run()
	if err != nil {
		return err
	}

	logger.Ologger.Debug(stdBuffer.String())

	logger.Ologger.Info(name + " template created")

	return nil
}

func (vm GoogleCompute) DeleteTemplate(name string) error {
	return vm.DeleteTemplateWait(name, true)
}

func (vm GoogleCompute) DeleteTemplateWait(name string, wait bool) error {

	ctx, templateClient, err := createTemplatesClient()
	if err != nil {
		return err
	}
	defer templateClient.Close()

	req := &computepb.DeleteInstanceTemplateRequest{
		InstanceTemplate: name,
		Project:          vm.project,
	}

	op, err := templateClient.Delete(ctx, req)
	if err != nil {
		return err
	}

	if wait {
		if err = op.Wait(ctx); err != nil {
			return err
		}
		logger.Ologger.Info(name + " template deleted")
	}

	return nil
}

func (vm GoogleCompute) ListTemplateInstances(filter string) error {

	ctx, templateClient, err := createTemplatesClient()
	if err != nil {
		return err
	}
	defer templateClient.Close()

	name := fmt.Sprint("name =", filter)

	req := &computepb.ListInstanceTemplatesRequest{
		Project: vm.project,
		Filter:  &name,
	}

	templates := templateClient.List(ctx, req)

	template, err := templates.Next()
	if err != nil {
		return err
	} else {
		logger.Ologger.Info("Found template " + *template.Name)
	}

	for template, err := templates.Next(); err == nil; template, err = templates.Next() {
		logger.Ologger.Info("Found template " + *template.Name)
	}

	return nil
}

func (vm GoogleCompute) GetTemplateInstance(filter string) (*computepb.InstanceTemplate, error) {

	ctx, templateClient, err := createTemplatesClient()
	if err != nil {
		return nil, err
	}
	defer templateClient.Close()

	name := fmt.Sprint("name =", filter)

	req := &computepb.ListInstanceTemplatesRequest{
		Project: vm.project,
		Filter:  &name,
	}

	templates := templateClient.List(ctx, req)

	template, err := templates.Next()
	if err != nil {
		return nil, err
	} else {
		return template, nil
	}
}
