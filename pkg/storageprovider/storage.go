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

// Topology contains the entire topology of the storage system
type Topology struct {
	/* TBD */
}

// Interface is a pluggable interface for storage providers
type Interface interface {
	// Topology returns the current topology and utilization of the
	// storage system
	GetTopology() (*Topology, error)

	// Utilization returns the total utilization of the storage system
	Utilization() (int, error)

	// DeviceAdd notifies the storage provider a new device has been added
	DeviceAdd( /* TBD */ ) error

	// DeviceRemove requests to remove a device from the storage system
	DeviceRemove( /* TBD */ ) error

	// Event handler TBD
	Event( /* TBD */ )
}
