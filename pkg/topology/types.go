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

// DeviceMetadata contains cloud metadata for the device
type DeviceMetadata struct {
	// Cloud volume id for this device
	ID string
}

// Device contains information about the device of the storage system
type Device struct {
	// Path of the block device node
	Path string

	// Class of device
	Class string

	// Pool name
	Pool string

	// Size in GiB
	Size int64

	// Utilization of the device as a percentage number
	Utilization int

	// Metadata has cloud identification for the device
	Metadata DeviceMetadata

	// Private can be used by the storage system as a cookie
	Private interface{}
}

// Pool contains a set of devices
type Pool struct {
	// Name or ID of the pool if any
	Name string

	// SetSize is the number of disks that should be added
	// or removed from the pool
	SetSize int

	// Utilization of the pool according to the storage system. The storage
	// provider must supply this information according their pool implementation
	Utilization int

	// Class of device used
	Class string

	// Private can be used by the storage system as a cookie
	Private interface{}
}

// InstanceMetadata contains cloud information about the instance
type InstanceMetadata struct {
	// ID is the cloud instance ID
	ID string

	// Zone holds cloud failure domain information
	Zone string
}

// StorageNode defines information about the node
type StorageNode struct {
	// Name/ID is the name of the node according to the storage system
	Name string

	// Metadata is the cloud information about this instance
	Metadata InstanceMetadata

	// Devices is a list of devices on this node
	Devices []*Device

	// Pool of devices on the node. Keys are the names of the pool
	Pools map[string]*Pool

	// Classes is a list of classes supported. If none are provided,
	// it defaults to all
	Classes []string

	// Private can be used by the storage system as a cookie
	Private interface{}
}

// StorageCluster is a collection of nodes
type StorageCluster struct {
	// StorageNodes is a list of nodes on this cluster
	StorageNodes []*StorageNode

	// Private can be used by the storage system as a cookie
	Private interface{}
}

// Topology contains the entire topology of the storage system
type Topology struct {
	Cluster StorageCluster
}
