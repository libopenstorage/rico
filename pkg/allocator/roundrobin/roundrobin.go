/*
Package roundrobin test test test
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
	"fmt"

	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/topology"
)

// Allocator provides a simple generic allocator for Rico
type Allocator struct{}

// New returns a new RoundRobinAllocator
func New() *Allocator {
	return &Allocator{}
}

// DetermineNodeToAddStorage returns a node on the cluster which will be used to add storage
// TODO: This will be an inteface to a new algorithm object
func (r *Allocator) DetermineNodeToAddStorage(
	t *topology.Topology,
	class *config.Class,
) (*topology.StorageNode, error) {
	if len(t.Cluster.StorageNodes) == 0 {
		return nil, fmt.Errorf("No storage nodes in the cluster")
	}
	node := t.Cluster.StorageNodes[0]
	for _, currentNode := range t.Cluster.StorageNodes {
		if len(currentNode.Devices) < len(node.Devices) {
			node = currentNode
		}
	}

	return node, nil
}

// DetermineStorageToRemove returns device to remove
// TODO: This will be an inteface to a new algorithm object
func (r *Allocator) DetermineStorageToRemove(
	t *topology.Topology,
	class *config.Class,
) (*topology.StorageNode, *topology.Pool, *topology.Device) {
	var (
		node   *topology.StorageNode
		device *topology.Device
		pool   *topology.Pool
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
