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

	"go.pedge.io/dlog"

	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/storageprovider"
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

// Start starts the eventloop
func (m *Manager) Start() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.running {
		return fmt.Errorf("already running")
	}

	m.running = true
	m.quit = make(chan struct{})

	// Start the eventloop
	started := make(chan bool)
	go m.eventloop(started)
	<-started
	return nil
}

// Stop stops the eventloop
func (m *Manager) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	close(m.quit)
	m.running = false
}

// IsRunning returns true if the eventloop is running
func (m *Manager) IsRunning() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.running
}

// SetConfig saves a new configuration value
func (m *Manager) SetConfig(config *config.Config) {
	// TODO - LOCK
	m.config = *config
}

// TODO:
// Create a simple example of a loop outside this package
func (m *Manager) eventloop(started chan<- bool) {
	dlog.Infoln("Started loop")
	started <- true

	// Wait to be told when to reconcile
	for {
		select {
		case <-m.quit:
			dlog.Infoln("Stopped loop")
			return
		case <-m.reconcile:
			if err := m.do(); err != nil {
				dlog.Errorln("%v", err)
			}
		}
	}
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
			dlog.Infof("Adding storage")
			return m.addStorage(t, &class)
		}

		if (utilization <= class.WatermarkLow &&
			totalStorage-class.DiskSizeGb >= class.MinimumTotalSizeGb) ||
			totalStorage > class.MaximumTotalSizeGb {
			dlog.Infof("remove storage")
			return m.removeStorage(t, &class)
		}
		dlog.Infof("No change")
	}
	return nil
}

func (m *Manager) addStorage(t *storageprovider.Topology, class *config.Class) error {
	// Pick a node
	node := t.DetermineNodeToAddStorage()

	// Determine how many disks we need to add to this node
	// TODO: NumDisks to be added
	numDisks, p := node.NumDisks(class)

	// Add disks to the node
	devices := make([]*storageprovider.Device, 0)
	for d := 0; d < numDisks; d++ {
		// Create and attach a disk to the node
		device, err := m.cloud.DeviceCreate(node.Metadata.ID, class)
		if err != nil {
			return fmt.Errorf("Failed to add disk to node %s: %v",
				node.Metadata.ID,
				err)
		}
		devices = append(devices, &storageprovider.Device{
			Class: class.Name,
			Path:  device.Path,
			Size:  device.Size,
			Metadata: storageprovider.DeviceMetadata{
				ID: device.ID,
			}})
	}

	// Notify storage system device has been added
	// TODO: Clean up on error
	return m.storage.DeviceAdd(node, p, devices)
}

func (m *Manager) removeStorage(t *storageprovider.Topology, class *config.Class) error {
	// Pick a device
	node, pool, device := t.DetermineStorageToRemove(class)

	// Nothing to do
	if device == nil {
		dlog.Infof("No device found to remove")
		return nil
	}

	// Remove drive from the storage system
	cloudDevices, err := m.storage.DeviceRemove(node, pool, device)
	if err != nil {
		return err
	}

	// Delete cloud drive
	var deleteErr error
	for _, d := range cloudDevices {
		err = m.cloud.DeviceDelete(node.Metadata.ID, device.Metadata.ID)
		if err != nil {
			deleteErr = err
			dlog.Errorf("Failed to remove cloud device %s: %v",
				d.Metadata.ID, err)
		}
	}
	return deleteErr
}
