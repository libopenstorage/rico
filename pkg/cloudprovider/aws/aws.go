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

	"github.com/libopenstorage/logrus"
	awsops "github.com/libopenstorage/openstorage/pkg/storageops/aws"
	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
)

// Provider has the client and state information to communicate with AWS
type Provider struct {
	ec2c *ec2.EC2
}

// NewProvider provides an implementation of cloudprovider.Instance
func NewProvider() *Provider {

	// https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html
	region := os.Getenv("AWS_DEFAULT_REGION")
	if len(region) == 0 {
		logrus.Errorf("AWS_DEFAULT_REGION not defined")
		return nil
	}

	// Create a session
	// TODO: When running in Kubernetes, look at how they do it.
	ec2c := ec2.New(session.New(
		&aws.Config{
			Region: &region,
		},
	))
	if ec2c == nil {
		return nil
	}

	return &Provider{
		ec2c: ec2c,
	}
}

// SetConfig does nothing
func (p *Provider) SetConfig(config *config.Config) {}

func (p *Provider) volumeRequestFromParameters(
	class *config.Class,
	volreq *ec2.Volume,
) error {

	// Get parameters
	// TODO: Get parameters
	voltype := opsworks.VolumeTypeGp2
	volreq.VolumeType = &voltype
	return nil
}

// DeviceCreate creates and attaches a device to a specific node
func (p *Provider) DeviceCreate(
	instanceID string,
	class *config.Class,
) (*cloudprovider.Device, error) {

	// Create an aws ops object
	ops := awsops.NewEc2Storage(instanceID, p.ec2c)

	// Get availability zone of the instance
	descriptionI, err := ops.Describe()
	description := descriptionI.(*ec2.Instance)
	if err != nil {
		return nil, err
	}

	// Create a volume request according to the parameters
	az := description.Placement.AvailabilityZone

	// Create a volume
	volreq := &ec2.Volume{
		AvailabilityZone: az,
		Size:             &class.DiskSizeGb,
	}
	if err = p.volumeRequestFromParameters(class, volreq); err != nil {
		return nil, err
	}

	d, err := ops.Create(volreq, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create volume: %v", err)
	}
	vol := d.(*ec2.Volume)

	// Attach the volume
	path, err := ops.Attach(*vol.VolumeId)
	if err != nil {
		reterr := fmt.Errorf("Unable to attach volume %s to %s: %v",
			*vol.VolumeId,
			instanceID,
			err)
		logrus.Errorf(err.Error())
		if err := ops.Delete(*vol.VolumeId); err != nil {
			logrus.Errorf("Failed to delete volume %s: %v", *vol.VolumeId, err)
		}
		return nil, reterr
	}

	return &cloudprovider.Device{
		ID:   *vol.VolumeId,
		Path: path,
		Size: *vol.Size,
	}, nil
}

// DeviceDelete detaches the volume from the specified node, then deletes it
func (p *Provider) DeviceDelete(instanceID string, deviceID string) error {
	// Create an aws ops object
	ops := awsops.NewEc2Storage(instanceID, p.ec2c)

	// Detach volume
	if err := ops.Detach(deviceID); err != nil {
		return fmt.Errorf("Failed to detach volume %s from instance %s: %v",
			deviceID,
			instanceID,
			err)
	}

	// Delete volume
	if err := ops.Delete(deviceID); err != nil {
		return fmt.Errorf("Failed to delete volume %s: %v",
			deviceID,
			err)
	}

	return nil
}
