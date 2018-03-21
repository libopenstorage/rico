/*
Package allocator provides the interface and implementations of
an allocator to Rico.
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
package allocator

import (
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/topology"
)

// Interface provides an algorithm to determine where to add and remove storage
type Interface interface {

	// DetermineStorageToRemove returns which device in the topology to remove
	DetermineStorageToRemove(*topology.Topology, *config.Class) (*topology.StorageNode, *topology.Pool, *topology.Device)

	// DetermineNodeToAddStorage returns a node which storage can be added to
	DetermineNodeToAddStorage(*topology.Topology, *config.Class) (*topology.StorageNode, error)
}
