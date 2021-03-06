/*
Package cloudprovider provides the interfaces to the cloud provider
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
//go:generate mockgen -package=mock -destination=mock/cloud.mock.go github.com/libopenstorage/rico/pkg/cloudprovider Interface
package cloudprovider

import (
	"github.com/libopenstorage/rico/pkg/config"
)

// Device container generic cloud information
type Device struct {

	// Cloud device ID
	ID string

	// Path of device node on the host
	Path string

	// Size in GiB
	Size int64
}

// Interface provides a pluggable interface for cloud providers
type Interface interface {
	// SetConfig updates the stroageprovider with the configuration provided
	SetConfig(config *config.Config)

	// DeviceAdd creates and attaches new device to a node returning
	// the id of the newly created device
	DeviceCreate(instanceID string, class *config.Class) (*Device, error)

	// DeviceDelete detaches and deletes a cloud block device from a node
	DeviceDelete(instanceID, deviceID string) error
}
