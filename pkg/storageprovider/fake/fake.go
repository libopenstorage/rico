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
	"github.com/libopenstorage/rico/pkg/storageprovider"
	"github.com/lpabon/godbc"
)

// This is for tests only

// Fake is an memory-only implementation of the storageprovider.Interface
type Fake struct {
	CurrentUtilization int
	Topology           *storageprovider.Topology
}

// New returns a new Fake storage implementation
func New(t *storageprovider.Topology) *Fake {
	return &Fake{
		Topology:           t,
		CurrentUtilization: 50,
	}
}

// GetTopology returns the topology kept in memory
func (f *Fake) GetTopology() (*storageprovider.Topology, error) {
	return f.Topology, nil
}

// Utilization retuns the system current utilization
func (f *Fake) Utilization() (int, error) {
	return f.CurrentUtilization, nil
}

// DeviceAdd adds a device to the topology
func (f *Fake) DeviceAdd(
	node *storageprovider.StorageNode,
	device *storageprovider.Device,
) error {

	found := false
	for _, sn := range f.Topology.Cluster.StorageNodes {
		if sn.Metadata.ID == node.Metadata.ID {
			found = true
			if len(sn.Devices) == 0 {
				sn.Devices = []*storageprovider.Device{device}
			} else {
				sn.Devices = append(sn.Devices, device)
			}
			break
		}
	}
	godbc.Ensure(found == true)
	return nil
}

// DeviceRemove removes a device from the topology
func (f *Fake) DeviceRemove(
	node *storageprovider.StorageNode,
	device *storageprovider.Device,
) error {
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
	return nil
}

// NumDevices returns the total number of devices on the Fake storage cluster
func (f *Fake) NumDevices() int {
	devices := 0
	for _, n := range f.Topology.Cluster.StorageNodes {
		devices += len(n.Devices)
	}
	return devices
}

// Event :TODO:
func (f *Fake) Event() {

}
