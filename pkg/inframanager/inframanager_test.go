/*
Package inframanager provides an interface to the infrastrcture manager
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
package inframanager

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/rico/pkg/storageprovider"

	"github.com/libopenstorage/rico/pkg/cloudprovider/aws"
	"github.com/libopenstorage/rico/pkg/storageprovider/fake"
)

func TestWithAws(t *testing.T) {
	inputInstances := os.Getenv("RICO_TEST_INSTANCES")
	if len(inputInstances) == 0 {
		t.Skipf("Must provide a comma separated list of instance ids in the " +
			"environment variable RICO_TEST_INSTANCES")
	}
	instances := strings.Split(inputInstances, ",")
	numInstances := len(instances)

	// Create a topology
	nodes := make([]*storageprovider.StorageNode, numInstances)
	for i, instance := range instances {
		nodes[i] = &storageprovider.StorageNode{
			Metadata: storageprovider.InstanceMetadata{
				ID: instance,
			},
		}
	}
	topology := &storageprovider.Topology{
		Cluster: storageprovider.StorageCluster{
			StorageNodes: nodes,
		},
	}

	// Create providers
	storage := fake.New(topology)
	cloud := aws.NewProvider()

	// Create a new manager
	im := NewManager(&Config{
		watermarkHigh: 75,
		watermarkLow:  25,
		diskSets:      1,
		diskSizeGb:    8,
	}, cloud, storage)
	assert.NotNil(t, im)

	// Start with a high watermark
	assert.Equal(t, 0, storage.NumDevices())
	storage.CurrentUtilization = 80
	for i := 0; i < (2 * numInstances); i++ {
		err := im.do()
		assert.NoError(t, err)
		assert.Equal(t, i+1, storage.NumDevices())
	}

	// Not above or below watermark, so there should be
	// no changes to the devices
	storage.CurrentUtilization = 50
	for i := 0; i < (2 * numInstances); i++ {
		err := im.do()
		assert.NoError(t, err)
		assert.Equal(t, 2*numInstances, storage.NumDevices())
	}

	// Low watermark tests
	storage.CurrentUtilization = 10
	for i := 0; i < (2 * numInstances); i++ {
		err := im.do()
		assert.NoError(t, err)
		assert.Equal(t, 2*numInstances-(i+1), storage.NumDevices())
	}
}
