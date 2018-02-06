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
	"time"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/storageprovider"
)

// Config contains all the configuration settings
type Config struct {

	// Add storage if utilization is above this value
	watermarkHigh int

	// Remove storage if utilization is below this value
	watermarkLow int

	// Add this many devices at a time. This is useful for systems
	// which support multiple replicas
	diskSets int

	// Size of the disk to add
	// TODO: This may be adjusted in future changes
	diskSizeGb uint64
}

// Manager is an implementation of inframanager.Interface
type Manager struct {
	config  Config
	lock    sync.Mutex
	running bool
	quit    chan struct{}
	cloud   cloudprovider.Interface
	storage storageprovider.Interface
}

// NewManager returns a new infrastructure manager implementation
func NewManager(
	config *Config,
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

func (m *Manager) eventloop(started chan<- bool) {
	dlog.Infoln("Started loop")
	started <- true

	// This event loop is JUST a place holder. This is
	// still under development.
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-m.quit:
			dlog.Infoln("Stopped loop")
			return
		case <-ticker.C:
			if err := m.do(); err != nil {
				dlog.Errorln("%v")
			}
		}
	}
}

func (m *Manager) do() error {
	utilization, err := m.storage.Utilization()
	if err != nil {
		return fmt.Errorf("Failed to get utilization: %v", err)
	}

	if utilization > m.config.watermarkHigh {
		return m.addStorage()
	} else if utilization < m.config.watermarkLow {
		return m.removeStorage()
	}

	return nil
}

func (m *Manager) addStorage() error {
	t, err := m.storage.GetTopology()
	if err != nil {
		return fmt.Errorf("Failed to get topology: %v", err)
	}

	if len(t.Cluster.StorageNodes) == 0 {
		return fmt.Errorf("Cluster has no storage nodes")
	}

	// TODO: Get nodes from multiple separate zones
	for set := 0; set < m.config.diskSets; set++ {
		// Pick a node
		// TODO: This will be an inteface to a new algorithm object
		node := t.Cluster.StorageNodes[0]
		for _, currentNode := range t.Cluster.StorageNodes {
			if len(currentNode.Devices) < len(node.Devices) {
				node = currentNode
			}
		}

		// Create and attach a disk to the node
		device, err := m.cloud.DeviceCreate(node.Metadata.ID, &cloudprovider.DeviceSpecs{
			Size: m.config.diskSizeGb,
		})
		if err != nil {
			return fmt.Errorf("Failed to add disk to node %s: %v",
				node.Metadata.ID,
				err)
		}

		// Notify storage system device has been added
		err = m.storage.DeviceAdd(node, &storageprovider.Device{
			Path: device.Path,
			Size: device.Size,
			Metadata: storageprovider.DeviceMetadata{
				ID: device.ID,
			},
		})
	}

	return nil
}

func (m *Manager) removeStorage() error {
	t, err := m.storage.GetTopology()
	if err != nil {
		return fmt.Errorf("Failed to get topology: %v", err)
	}

	if len(t.Cluster.StorageNodes) == 0 {
		return fmt.Errorf("Cluster has no storage nodes")
	}

	// Pick a device
	// This is a silly algorithm for now
	// TODO: This will be an inteface to a new algorithm object
	var device *storageprovider.Device
	var node *storageprovider.StorageNode
	for _, currentNode := range t.Cluster.StorageNodes {
		for _, currentDevice := range currentNode.Devices {
			if device == nil {
				node = currentNode
				device = currentDevice
			} else if currentDevice.Utilization < device.Utilization {
				node = currentNode
				device = currentDevice
			}
		}
	}

	// Nothing to do
	if device == nil {
		return nil
	}

	// Remove drive from the storage system
	if err = m.storage.DeviceRemove(node, device); err != nil {
		return err
	}

	// Delete cloud drive
	return m.cloud.DeviceDelete(node.Metadata.ID, device.Metadata.ID)
}
