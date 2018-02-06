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
		{
			size:     4,
			instance: instances[3%numInstances],
		},
		{
			size:     12,
			instance: instances[4%numInstances],
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
		info, err := a.ops.DescribeInstance(test.instance)
		assert.Nil(t, err)
		assert.NotNil(t, info)

		found := false
		for _, bd := range info.BlockDeviceMappings {
			if *bd.Ebs.VolumeId == device.ID {
				found = true

				// TODO(lpabon)
				// The following will need to be investigated to see
				// what ec2ops is doing.
				/*
										        Test:           TestDeviceAddDelete
						        Error Trace:    aws_test.go:84
						        Error:          Not equal:
						                        expected: "/dev/sdf"
						                        actual  : "/dev/xvdf"
						        Test:           TestDeviceAddDelete
						        Error Trace:    aws_test.go:84
						        Error:          Not equal:
						                        expected: "/dev/sdg"
												actual  : "/dev/xvdg"
					assert.Equal(t, *bd.DeviceName, device.Path)
				*/
			}
		}
		assert.True(t, found)

		// Delete device
		err = a.DeviceDelete(test.instance, device.ID)
		assert.Nil(t, err)

		// Check that the device was added
		info, err = a.ops.DescribeInstance(test.instance)
		assert.Nil(t, err)
		assert.NotNil(t, info)

		found = false
		for _, bd := range info.BlockDeviceMappings {
			if *bd.Ebs.VolumeId == device.ID {
				found = true
			}
		}
		assert.False(t, found)
	}
}
