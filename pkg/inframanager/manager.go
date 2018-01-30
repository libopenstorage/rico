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

// Config provides configuration to the InfraManager implementation
type Config struct {
	/* TBD */
}

// Manager is the interface to an infrastructure manager implementation
type Manager interface {

	// Configure sets the configuration of the infrastructure manager
	Configure(*Config) error

	// Start the service using a specific configuration
	Start() error

	// Stop the InfraManager service
	Stop()

	// IsRunning returns true if the service is running
	IsRunning() bool
}
