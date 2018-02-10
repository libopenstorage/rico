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
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	awsops "github.com/libopenstorage/openstorage/pkg/storageops/aws"
	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/stretchr/testify/assert"
)

/*
 Create RICO_TEST_INSTANCES:

 export RICO_TEST_INSTANCES=$(aws ec2 describe-instances \
	 --filters Name=instance-state-name,Values=running | \
	 jq '.Reservations[].Instances[].InstanceId' | \
	 tr '\n' ' ' | \
	 sed -e "s#\"##g" -e "s# #,#g" -e 's#.$##')

*/

func TestAwsDeviceAddDelete(t *testing.T) {
	inputInstances := os.Getenv("RICO_TEST_INSTANCES")
	if len(inputInstances) == 0 {
		t.Skipf("Must provide a comma separated list of instance ids in the " +
			"environment variable RICO_TEST_INSTANCES")
	}
	instances := strings.Split(inputInstances, ",")
	numInstances := len(instances)

	a := NewProvider()
	assert.NotNil(t, a)

	tests := []struct {
		size     uint64
		instance string
	}{
		{
			size:     8,
			instance: instances[1%numInstances],
		},
		{
			size:     2,
			instance: instances[2%numInstances],
		},
	}

	// Add Devices
	for _, test := range tests {
		// Add a device to the instance
		device, err := a.DeviceCreate(test.instance, &cloudprovider.DeviceSpecs{
			Size: test.size,
		})
		assert.Nil(t, err)
		assert.NotNil(t, device)
		assert.NotEmpty(t, device.Path)
		assert.NotEmpty(t, device.ID)
		assert.Equal(t, device.Size, test.size)

		// Check that the device was added
		ops := awsops.NewEc2Storage(test.instance, a.ec2c)
		infoI, err := ops.Describe()
		info := infoI.(*ec2.Instance)

		assert.Nil(t, err)
		assert.NotNil(t, info)

		found := false
		for _, bd := range info.BlockDeviceMappings {
			if *bd.Ebs.VolumeId == device.ID {
				found = true
				assert.Equal(t, *bd.DeviceName, device.Path)
				break
			}
		}
		assert.True(t, found)

		// Delete device
		err = a.DeviceDelete(test.instance, device.ID)
		assert.Nil(t, err)

		// Check that the device was added
		infoI, err = ops.Describe()
		info = infoI.(*ec2.Instance)
		assert.Nil(t, err)
		assert.NotNil(t, info)

		found = false
		for _, bd := range info.BlockDeviceMappings {
			if *bd.Ebs.VolumeId == device.ID {
				found = true
				break
			}
		}
		assert.False(t, found)
	}
}
