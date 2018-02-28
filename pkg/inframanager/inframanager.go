/*
Package inframanager provides an interface to the infrastrcture manager
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
package inframanager

import (
	"fmt"
	"sync"

	"github.com/libopenstorage/logrus"
	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/storageprovider"
	"github.com/libopenstorage/rico/pkg/topology"
)

// Manager is an implementation of inframanager.Interface
type Manager struct {
	config    config.Config
	lock      sync.Mutex
	running   bool
	quit      chan struct{}
	reconcile chan struct{}
	cloud     cloudprovider.Interface
	storage   storageprovider.Interface
}

// NewManager returns a new infrastructure manager implementation
func NewManager(
	config *config.Config,
	cloud cloudprovider.Interface,
	storage storageprovider.Interface,
) *Manager {
	return &Manager{
		config:  *config,
		cloud:   cloud,
		storage: storage,
	}
}

// SetConfig saves a new configuration value
func (m *Manager) SetConfig(config *config.Config) {
	// TODO - LOCK
	m.config = *config
}

// Config returns a copy of the current configuration
func (m *Manager) Config() *config.Config {
	c := m.config
	return &c
}

// Reconcile adds or removes storage from the system
func (m *Manager) Reconcile() error {
	return m.do()
}

func (m *Manager) do() error {

	// Get topology from the storage system
	t, err := m.storage.GetTopology()
	if err != nil {
		return err
	}

	// Verify the topology was filled in correctly
	if err := t.Verify(); err != nil {
		return err
	}

	// Check the utilization of each class
	for _, class := range m.config.Classes {
		utilization := t.Utilization(&class)
		totalStorage := t.TotalStorage(&class)

		// Do not add any more storage if at the max
		if (utilization >= class.WatermarkHigh &&
			totalStorage+class.DiskSizeGb <= class.MaximumTotalSizeGb) ||
			totalStorage < class.MinimumTotalSizeGb {
			m.addStorage(t, &class)
		} else if (utilization <= class.WatermarkLow &&
			totalStorage-class.DiskSizeGb >= class.MinimumTotalSizeGb) ||
			totalStorage > class.MaximumTotalSizeGb {
			return m.removeStorage(t, &class)
		} else {
			logrus.Infof("class:%s No change", class.Name)
		}
	}
	return nil
}

func (m *Manager) addStorage(t *topology.Topology, class *config.Class) error {
	// Pick a node
	node := t.DetermineNodeToAddStorage()

	// Determine how many disks we need to add to this node
	// TODO: NumDisks to be added
	numDisks, p := node.SetSizeForClass(class)

	// Add disks to the node
	devices := make([]*topology.Device, 0)
	for d := 0; d < numDisks; d++ {
		logrus.Infof("class:%s Creating/attaching storage %d of %d to node:%s",
			class.Name,
			d,
			numDisks,
			node.Metadata.ID)
		// Create and attach a disk to the node
		device, err := m.cloud.DeviceCreate(node.Metadata.ID, class)
		if err != nil {
			return fmt.Errorf("Failed to add disk to node %s: %v",
				node.Metadata.ID,
				err)
		}
		devices = append(devices, &topology.Device{
			Class: class.Name,
			Path:  device.Path,
			Size:  device.Size,
			Metadata: topology.DeviceMetadata{
				ID: device.ID,
			}})
	}

	// Notify storage system device has been added
	// TODO: Clean up on error
	logrus.Infof("class:%s Notifying storage system addition of %d to node:%s",
		class.Name,
		numDisks,
		node.Metadata.ID)
	return m.storage.DeviceAdd(node, p, devices)
}

func (m *Manager) removeStorage(t *topology.Topology, class *config.Class) error {
	// Pick a device
	node, pool, device := t.DetermineStorageToRemove(class)

	// Nothing to do
	if device == nil {
		logrus.Infof("class:%s No device found to remove", class.Name)
		return nil
	}

	// Remove drive from the storage system
	logrus.Infof("class:%s Removing device %s/%s:%s from storage",
		class.Name,
		node.Metadata.ID,
		device.Path,
		device.Metadata.ID)
	cloudDevices, err := m.storage.DeviceRemove(node, pool, device)
	if err != nil {
		return err
	}

	// Delete cloud drive
	var deleteErr error
	for _, d := range cloudDevices {
		logrus.Infof("class:%s Detaching/deleting device %s/%s:%s from storage",
			class.Name,
			node.Metadata.ID,
			d.Path,
			d.Metadata.ID)
		err = m.cloud.DeviceDelete(node.Metadata.ID, device.Metadata.ID)
		if err != nil {
			deleteErr = err
			logrus.Errorf("Failed to remove cloud device %s: %v",
				d.Metadata.ID, err)
		}
	}
	return deleteErr
}
