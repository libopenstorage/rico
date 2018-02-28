/*
Package fake provides an in-memory implementation of the cloud interface
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
	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/config"

	"github.com/pborman/uuid"
)

// Fake is an in-memory fake cloud provider
type Fake struct{}

// New returns a new Fake cloud provider
func New() *Fake {
	return &Fake{}
}

// SetConfig does nothing
func (f *Fake) SetConfig(config *config.Config) {}

// DeviceCreate returns a new device with a new uuid
func (f *Fake) DeviceCreate(
	instanceID string,
	class *config.Class,
) (*cloudprovider.Device, error) {
	return &cloudprovider.Device{
		ID:   uuid.New(),
		Path: "nothing",
		Size: class.DiskSizeGb,
	}, nil
}

// DeviceDelete does nothing
func (f *Fake) DeviceDelete(instanceID, deviceID string) error {
	return nil
}
