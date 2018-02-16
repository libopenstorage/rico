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
package storageprovider

//
// THIS FILE WILL BECOME A new pkg/topology and each of the objects
// broken into a file in that package.
//

import (
	"fmt"

	"github.com/libopenstorage/rico/pkg/config"
)

// Utilization returns the average utilization for a specified class
// across the entire cluster.
func (t *Topology) Utilization(class *config.Class) int {
	sum := 0
	for _, node := range t.Cluster.StorageNodes {
		sum += node.Utilization(class)
	}
	// TODO Check for DivZero
	return int(sum / len(t.Cluster.StorageNodes))
}

func (n *StorageNode) Utilization(class *config.Class) int {
	sum, num := 0, 0
	if len(n.Pools) != 0 {
		for _, pool := range n.Pools {
			if class.Name == pool.Class {
				sum += pool.Utilization
				num++
			}
		}
	} else {
		for _, device := range n.Devices {
			if class.Name == device.Class {
				sum += device.Utilization
				num++
			}
		}
	}
	if num == 0 {
		return 0
	}
	return int(sum / num)
}

// TODO: Make Size an explicit type as int64
func (t *Topology) TotalStorage(class *config.Class) int64 {
	total := int64(0)
	for _, n := range t.Cluster.StorageNodes {
		total += n.TotalStorage(class)
	}
	return total
}

func (n *StorageNode) TotalStorage(class *config.Class) int64 {
	total := int64(0)
	for _, d := range n.Devices {
		if d.Class == class.Name {
			total += d.Size
		}
	}
	return total
}

// DetermineNodeToAddStorage returns a node on the cluster which will be used to add storage
// TODO: This will be an inteface to a new algorithm object
func (t *Topology) DetermineNodeToAddStorage() *StorageNode {
	node := t.Cluster.StorageNodes[0]
	for _, currentNode := range t.Cluster.StorageNodes {
		if len(currentNode.Devices) < len(node.Devices) {
			node = currentNode
		}
	}

	return node
}

// DetermineStorageToRemove returns device to remove
// TODO: This will be an inteface to a new algorithm object
func (t *Topology) DetermineStorageToRemove(
	class *config.Class,
) (*StorageNode, *Pool, *Device) {
	var (
		node    *StorageNode
		device  *Device
		pool    *Pool
		devices []*Device
	)

	// Get the node
	for _, currentNode := range t.Cluster.StorageNodes {
		if len(currentNode.Devices) == 0 {
			continue
		}
		if node == nil ||
			currentNode.Utilization(class) < node.Utilization(class) {
			node = currentNode
		}
	}
	if node == nil {
		return nil, nil, nil
	}

// BUG: Pool needs to check the class
	if len(node.Pools) != 0 {
		for _, currentpool := range node.Pools {
			if pool == nil {
				pool = currentpool
			} else if currentpool.Utilization < pool.Utilization {
				pool = currentpool
			}
		}

		// Pick devices in the pull
		devices = node.DevicesOnPool(pool)
	} else {
		devices = node.Devices
	}

	// Pick ONE drive, let the storage system figure out if it needs to
	// remove more
	for _, currentDevice := range devices {
		if device == nil ||
			currentDevice.Utilization < device.Utilization {
			device = currentDevice
		}
	}

	return node, pool, device
}

// Rename Function
func (n *StorageNode) NumDisks(class *config.Class) (int, *Pool) {
	var (
		numDisks int
		p        *Pool
		ok       bool
	)

	numDisks = 1
	if p, ok = n.Pools[class.Name]; ok {
		numDisks = p.SetSize
	}

	return numDisks, p
}

// Verify confirms that the topology has the information required
// TODO: This is not complete while this is WIP
func (t *Topology) Verify() error {
	if len(t.Cluster.StorageNodes) == 0 {
		return fmt.Errorf("No storage nodes available in cluster")
	}
	for _, node := range t.Cluster.StorageNodes {
		if err := node.Verify(); err != nil {
			return err
		}
	}

	return nil
}

func (n *StorageNode) Verify() error {
	if len(n.Metadata.ID) == 0 {
		return fmt.Errorf("Node missing instance metadata id")
	}
	for _, pool := range n.Pools {
		if err := pool.Verify(); err != nil {
			return err
		}
	}
	for _, device := range n.Devices {
		if err := device.Verify(); err != nil {
			return err
		}
	}

	return nil
}


// TODO: DeviceInPool
func (n *StorageNode) DevicesOnPool(p *Pool) []*Device {
	devices := make([]*Device, 0)
	for _, device := range n.Devices {
		if device.Pool == p.Name {
			devices = append(devices, device)
		}
	}

	return devices
}

func (p *Pool) Verify() error {
	if p.SetSize == 0 {
		return fmt.Errorf("Size in pool cannot be zero")
	}
	if len(p.Class) == 0 {
		return fmt.Errorf("Pool class type cannot be empty")
	}
	return nil
}

func (d *Device) Verify() error {
	if len(d.Metadata.ID) == 0 {
		return fmt.Errorf("Device metadata id cannot be zero")
	}
	if len(d.Class) == 0 {
		return fmt.Errorf("Device class type cannot be empty")
	}
	return nil
}

func (t *Topology) NumDevices() int {
	devices := 0
	for _, n := range t.Cluster.StorageNodes {
		devices += len(n.Devices)
	}
	return devices
}
