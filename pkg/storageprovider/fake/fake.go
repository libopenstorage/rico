/*
Package fake provides a fake implementation of the storage interface to
be used for testing.
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
package fake

import (
	"fmt"

	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/topology"
	"github.com/lpabon/godbc"
)

// This is for tests only

// Fake is an memory-only implementation of the topology.Interface
type Fake struct {
	Topology *topology.Topology
}

// New returns a new Fake storage implementation
func New(t *topology.Topology) *Fake {
	return &Fake{
		Topology: t,
	}
}

// SetConfig does nothing here
func (f *Fake) SetConfig(*config.Config) {}

// GetTopology returns the topology kept in memory
func (f *Fake) GetTopology() (*topology.Topology, error) {
	return f.Topology, nil
}

// SetUtilization sets all the nodes to the specified utilization
func (f *Fake) SetUtilization(
	class *config.Class,
	utilization int,
) {
	for _, n := range f.Topology.Cluster.StorageNodes {
		f.SetNodeUtilization(n, class, utilization)
	}
}

// SetNodeUtilization sets all the devices to a utilization
func (f *Fake) SetNodeUtilization(
	node *topology.StorageNode,
	class *config.Class,
	utilization int,
) {
	for _, d := range node.Devices {
		if d.Class == class.Name {
			d.Utilization = utilization
		}
	}
}

// DeviceAdd adds a device to the topology
func (f *Fake) DeviceAdd(
	node *topology.StorageNode,
	pool *topology.Pool,
	devices []*topology.Device,
) error {

	godbc.Require(pool == nil)

	found := false
	for _, sn := range f.Topology.Cluster.StorageNodes {
		if sn.Metadata.ID == node.Metadata.ID {
			found = true
			sn.Devices = append(sn.Devices, devices...)
			break
		}
	}

	godbc.Ensure(found == true)
	return nil
}

// NodeAdd adds a new node to the topology
func (f *Fake) NodeAdd(node *topology.StorageNode) error {
	f.Topology.Cluster.StorageNodes = append(f.Topology.Cluster.StorageNodes, node)
	return nil
}

// NodeDelete removes a node and all its devices from the topology
func (f *Fake) NodeDelete(instanceID string) error {
	index := 0
	found := false
	nodes := f.Topology.Cluster.StorageNodes
	for i, node := range nodes {
		if node.Metadata.ID == instanceID {
			found = true
			index = i
			break
		}
	}

	if !found {
		return fmt.Errorf("Instance %s not found", instanceID)
	}
	nodes[index] = nodes[len(nodes)-1]
	nodes = nodes[:len(nodes)-1]
	return nil
}

// DeviceRemove removes a device from the topology
func (f *Fake) DeviceRemove(
	node *topology.StorageNode,
	pool *topology.Pool,
	device *topology.Device,
) ([]*topology.Device, error) {
	found := false
	for _, sn := range f.Topology.Cluster.StorageNodes {
		if sn.Metadata.ID == node.Metadata.ID {
			index := 0
			for i, d := range sn.Devices {
				if d.Metadata.ID == device.Metadata.ID {
					found = true
					index = i
					break
				}
			}

			if found {
				sn.Devices[index] = sn.Devices[len(sn.Devices)-1]
				sn.Devices = sn.Devices[:len(sn.Devices)-1]
			}
		}
	}
	godbc.Ensure(found == true)
	return []*topology.Device{device}, nil
}
