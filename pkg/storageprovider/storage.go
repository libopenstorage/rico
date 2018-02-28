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

import (
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/topology"
)

// Interface is a pluggable interface for storage providers
type Interface interface {
	// SetConfig updates the storageprovider with the configuration provided
	SetConfig(*config.Config)

	// Topology returns the current topology and utilization of the storage system
	GetTopology() (*topology.Topology, error)

	// DeviceAdd notifies the storage provider a devices have been added
	// If Pool is nil if it is up to the storage system to decide which pool
	// to add it to if any.
	DeviceAdd(*topology.StorageNode, *topology.Pool, []*topology.Device) error

	// DeviceRemove requests to remove a device from the storage system.
	// The storage system must return a list of devices to then remove
	// from the infrastructure.
	DeviceRemove(*topology.StorageNode, *topology.Pool, *topology.Device) ([]*topology.Device, error)
}
