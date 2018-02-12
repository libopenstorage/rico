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
//go:generate mockgen -package=mock -destination=mock/storage.mock.go github.com/libopenstorage/rico/pkg/storageprovider Interface
package storageprovider

// DeviceMetadata contains cloud metadata for the device
type DeviceMetadata struct {
	// Cloud volume id for this device
	ID string
}

// Device contains information about the device of the storage system
type Device struct {
	// Path of the block device node
	Path string

	// Size in GiB
	Size uint64

	// Utilization of the device as a percentage number
	Utilization int

	// Metadata has cloud identification for the device
	Metadata DeviceMetadata

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

// Interface is a pluggable interface for storage providers
type Interface interface {
	// Topology returns the current topology and utilization of the storage system
	GetTopology() (*Topology, error)

	// Utilization returns the total utilization of the storage system
	Utilization() (int, error)

	// DeviceAdd notifies the storage provider a new device has been added
	DeviceAdd(*StorageNode, *Device) error

	// DeviceRemove requests to remove a device from the storage system
	DeviceRemove(*StorageNode, *Device) error

	// Event handler TBD
	Event( /* TBD */ )
}
