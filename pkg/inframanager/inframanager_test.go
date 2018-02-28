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

	"github.com/libopenstorage/rico/pkg/cloudprovider/aws"
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/storageprovider/fake"
	"github.com/libopenstorage/rico/pkg/topology"
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
	nodes := make([]*topology.StorageNode, numInstances)
	for i, instance := range instances {
		nodes[i] = &topology.StorageNode{
			Metadata: topology.InstanceMetadata{
				ID: instance,
			},
		}
	}
	topology := &topology.Topology{
		Cluster: topology.StorageCluster{
			StorageNodes: nodes,
		},
	}

	// Create providers
	storage := fake.New(topology)
	cloud := aws.NewProvider()
	class := config.Class{
		Name:               "gp2",
		WatermarkHigh:      75,
		WatermarkLow:       25,
		DiskSizeGb:         8,
		MaximumTotalSizeGb: 1024,
		MinimumTotalSizeGb: 32,
	}
	config := &config.Config{
		Classes: []config.Class{class},
	}

	// Create a new manager
	im := NewManager(config, cloud, storage)
	assert.NotNil(t, im)

	// Fill topology with disks
	topology, _ = storage.GetTopology()
	assert.Equal(t, 0, topology.NumDevices())
	loops := int(class.MinimumTotalSizeGb/class.DiskSizeGb) * 3
	for i := 0; i < loops; i++ {
		err := im.do()
		assert.NoError(t, err)
	}
	topology, _ = storage.GetTopology()
	numDevices := topology.NumDevices()
	assert.Equal(t, int(class.MinimumTotalSizeGb/class.DiskSizeGb), numDevices)

	// Start with a high watermark
	for i := 0; i < (2 * numInstances); i++ {
		storage.SetUtilization(&class, 80)
		err := im.do()
		assert.NoError(t, err)
		topology, _ = storage.GetTopology()
		assert.Equal(t, numDevices+i+1, topology.NumDevices())
	}
	topology, _ = storage.GetTopology()
	numDevices = topology.NumDevices()

	// Not above or below watermark, so there should be
	// no changes to the devices
	for i := 0; i < loops; i++ {
		storage.SetUtilization(&class, 50)
		err := im.do()
		assert.NoError(t, err)

		topology, _ = storage.GetTopology()
		assert.Equal(t, numDevices, topology.NumDevices())
	}

	// Low watermark tests
	for i := 0; i < numDevices; i++ {
		storage.SetUtilization(&class, 10)
		err := im.do()
		assert.NoError(t, err)

		topology, _ := storage.GetTopology()
		assert.True(t, topology.TotalStorage(&class) >= class.MinimumTotalSizeGb)
	}
	topology, _ = storage.GetTopology()
	numDevices = topology.NumDevices()
	assert.Equal(t, int(class.MinimumTotalSizeGb/class.DiskSizeGb), numDevices)

	// Delete all volumes
	im.config.Classes[0].MinimumTotalSizeGb = 0
	for i := 0; i < loops; i++ {
		storage.SetUtilization(&class, 0)
		topology, _ = storage.GetTopology()
		numDevices = topology.NumDevices()

		err := im.do()
		assert.NoError(t, err)

		topology, _ = storage.GetTopology()
		if numDevices != 0 {
			assert.Equal(t, numDevices-1, topology.NumDevices())
		} else {
			assert.Equal(t, 0, topology.NumDevices())
		}
	}
}
