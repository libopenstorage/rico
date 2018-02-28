/*
Package topology defines how to get information from the infrastructure
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
package topology

import (
	"fmt"

	"github.com/libopenstorage/rico/pkg/config"
)

// Utilization returns the average utilization for a specified class
// across the entire cluster.
func (t *Topology) Utilization(class *config.Class) int {
	sum := 0
	num := 0
	for _, node := range t.Cluster.StorageNodes {
		s, n := node.RawUtilization(class)
		sum += s
		num += n
	}
	// TODO Check for DivZero
	if num == 0 {
		return 0
	}
	return int(sum / num)
}

// TotalStorage returns the total storage in a topology allocated by a certain class
// TODO: Make Size an explicit type as int64
func (t *Topology) TotalStorage(class *config.Class) int64 {
	total := int64(0)
	for _, n := range t.Cluster.StorageNodes {
		total += n.TotalStorage(class)
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
		node   *StorageNode
		device *Device
		pool   *Pool
	)

	// Get the node
	for _, currentNode := range t.Cluster.StorageNodes {
		devices := currentNode.DevicesForClass(class)
		if len(devices) == 0 {
			continue
		}

		// Check pools on this node
		if len(currentNode.Pools) != 0 {
			for _, currentpool := range node.Pools {
				if currentpool.Class == class.Name {
					if pool == nil ||
						currentpool.Utilization < pool.Utilization {
						pool = currentpool
					}
				}
			}

			// Pick devices in the pull
			devices = node.DevicesOnPool(pool)
		}

		for _, currentDevice := range devices {
			if device == nil ||
				currentDevice.Utilization < device.Utilization {
				node = currentNode
				device = currentDevice
			}
		}

	}
	if node == nil {
		return nil, nil, nil
	}

	return node, pool, device
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

// NumDevices returns the total number of devices in the topolgy
func (t *Topology) NumDevices() int {
	devices := 0
	for _, n := range t.Cluster.StorageNodes {
		devices += len(n.Devices)
	}
	return devices
}

// String returns a string representation of the topology for fmt.Printf
func (t *Topology) String() string {
	s := ""
	for _, node := range t.Cluster.StorageNodes {
		s += node.String()
	}
	return s
}
