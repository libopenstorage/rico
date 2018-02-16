/*
Package config provides the configuration to the Manager
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
package config

// Class defines the type of storage to use for the appropriate
// cloud provider
// TODO: Use json instead
type Class struct {
	// Name of the class
	Name string `yaml:"name"`

	// Parameters for this class
	Parameters map[string]string `yaml:"parameters"`

	// Add storage if utilization is above this value
	WatermarkHigh int `yaml:"watermarkHigh"`

	// Remove storage if utilization is below this value
	WatermarkLow int `yaml:"watermarkLow"`

	// Maximum size in Gi of storage of this class on the cluster
	MaximumTotalSizeGb int64 `yaml:"maximumTotalSize"`

	// Minimum size in Gi of storage of this class on the cluster
	MinimumTotalSizeGb int64 `yaml:"minimumTotalSize"`

	// Size of the disk to add
	DiskSizeGb int64 `yaml:"diskSize"`
}

// Config contains all the configuration settings
type Config struct {

	// Classes of storage to manage
	Classes []Class `yaml:"classes"`
}
