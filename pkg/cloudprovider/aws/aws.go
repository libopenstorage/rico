/*
Package aws implements the cloud interface for AWS
Copyright 2018 Portworx

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
package aws

import (
	"fmt"
	"os"

	"github.com/libopenstorage/rico/pkg/cloudprovider"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"go.pedge.io/dlog"
)

// Provider has the client and state information to communicate with AWS
type Provider struct {
	ops  *Ec2Ops
	ec2c *ec2.EC2
}

// NewProvider provides an implementation of cloudprovider.Instance
func NewProvider() *Provider {

	// https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html
	region := os.Getenv("AWS_DEFAULT_REGION")
	if len(region) == 0 {
		dlog.Errorf("AWS_DEFAULT_REGION not defined")
		return nil
	}

	// Create a session
	// TODO: When running in Kubernetes, look at how they do it.
	ec2c := ec2.New(session.New(
		&aws.Config{
			Region: &region,
		},
	))
	ec2ops := NewEc2Ops(ec2c)

	return &Provider{
		ec2c: ec2c,
		ops:  ec2ops,
	}
}

// DeviceCreate creates and attaches a device to a specific node
func (p *Provider) DeviceCreate(
	instanceID string,
	device *cloudprovider.DeviceSpecs,
) (*cloudprovider.Device, error) {

	// Get availability zone of the instance
	description, err := p.ops.DescribeInstance(instanceID)
	if err != nil {
		return nil, err
	}
	az := description.Placement.AvailabilityZone

	// Create a volume
	size := int64(device.Size)
	voltype := opsworks.VolumeTypeGp2
	volreq := &ec2.Volume{
		AvailabilityZone: az,
		VolumeType:       &voltype,
		Size:             &size,
	}
	d, err := p.ops.Create(volreq, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create volume: %v", err)
	}
	vol, ok := d.(*ec2.Volume)
	if !ok {
		panic("Create returned an unexpected type")
	}

	// Attach the volume
	path, err := p.ops.Attach(instanceID, *vol.VolumeId)
	if err != nil {
		reterr := fmt.Errorf("Unable to attach volume %s to %s: %v",
			*vol.VolumeId,
			instanceID,
			err)
		dlog.Errorf(err.Error())
		if err := p.ops.Delete(*vol.VolumeId); err != nil {
			dlog.Errorf("Failed to delete volume %s: %v", *vol.VolumeId, err)
		}
		return nil, reterr
	}

	return &cloudprovider.Device{
		ID:   *vol.VolumeId,
		Path: path,
		Size: uint64(*vol.Size),
	}, nil
}

// DeviceDelete detaches the volume from the specified node, then deletes it
func (p *Provider) DeviceDelete(instanceID string, deviceID string) error {

	// Detach volume
	if err := p.ops.Detach(instanceID, deviceID); err != nil {
		return fmt.Errorf("Failed to detach volume %s from instance %s: %v",
			deviceID,
			instanceID,
			err)
	}

	// Delete volume
	if err := p.ops.Delete(deviceID); err != nil {
		return fmt.Errorf("Failed to delete volume %s: %v",
			deviceID,
			err)
	}

	return nil
}