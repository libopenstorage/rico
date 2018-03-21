/*
Package storageprovider provides an interface to storage providers
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
package roundrobin

import (
	"testing"

	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/topology"
	"github.com/stretchr/testify/assert"
)

func TestRRDetermineStorageToRemove(t *testing.T) {
	testTopology := &topology.Topology{
		Cluster: topology.StorageCluster{
			StorageNodes: []*topology.StorageNode{
				&topology.StorageNode{
					Metadata: topology.InstanceMetadata{
						ID: "one",
					},
					Devices: []*topology.Device{
						&topology.Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: topology.DeviceMetadata{
								ID: "d1",
							},
						},
						&topology.Device{
							Class:       "c1",
							Utilization: 1,
							Metadata: topology.DeviceMetadata{
								ID: "d2",
							},
						},
						&topology.Device{
							Class:       "c2",
							Utilization: 4,
							Metadata: topology.DeviceMetadata{
								ID: "d3",
							},
						},
						&topology.Device{
							Class:       "c2",
							Utilization: 3,
							Metadata: topology.DeviceMetadata{
								ID: "d4",
							},
						},
					},
				},
				&topology.StorageNode{
					Metadata: topology.InstanceMetadata{
						ID: "two",
					},
					Devices: []*topology.Device{
						&topology.Device{
							Class:       "c3",
							Utilization: 3,
							Metadata: topology.DeviceMetadata{
								ID: "d1",
							},
						},
						&topology.Device{
							Class:       "c3",
							Utilization: 3,
							Metadata: topology.DeviceMetadata{
								ID: "d2",
							},
						},
						&topology.Device{
							Class:       "c3",
							Utilization: 3,
							Metadata: topology.DeviceMetadata{
								ID: "d3",
							},
						},
						&topology.Device{
							Class:       "c3",
							Utilization: 3,
							Metadata: topology.DeviceMetadata{
								ID: "d4",
							},
						},
					},
				},
				&topology.StorageNode{
					Metadata: topology.InstanceMetadata{
						ID: "three",
					},
					Devices: []*topology.Device{
						&topology.Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: topology.DeviceMetadata{
								ID: "d1",
							},
						},
						&topology.Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: topology.DeviceMetadata{
								ID: "d2",
							},
						},
						&topology.Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: topology.DeviceMetadata{
								ID: "d3",
							},
						},
						&topology.Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: topology.DeviceMetadata{
								ID: "d4",
							},
						},
					},
				},
			},
		},
	}

	rr := New()
	n, p, d := rr.DetermineStorageToRemove(testTopology, &config.Class{
		Name: "c2",
	})
	assert.NotNil(t, n)
	assert.Equal(t, "one", n.Metadata.ID)
	assert.Nil(t, p)
	assert.NotNil(t, d)
	assert.Equal(t, "d4", d.Metadata.ID)
}
