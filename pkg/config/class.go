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

import (
	"fmt"
)

// Class defines the type of storage to use for the appropriate
// cloud provider
// TODO: Use json instead
type Class struct {
	// Name of the class
	Name string `json:"name"`

	// Parameters for this class
	Parameters map[string]string `json:"parameters"`

	// Add storage if utilization is above this value
	WatermarkHigh int `json:"watermarkHigh"`

	// Remove storage if utilization is below this value
	WatermarkLow int `json:"watermarkLow"`

	// Maximum size in Gi of storage of this class on the cluster
	MaximumTotalSizeGb int64 `json:"maximumTotalSize"`

	// Minimum size in Gi of storage of this class on the cluster
	MinimumTotalSizeGb int64 `json:"minimumTotalSize"`

	// Size of the disk to add
	DiskSizeGb int64 `json:"diskSize"`
}

// String formats a string based on the information from the class
func (c Class) String() string {
	return fmt.Sprintf("%s: Max:%d Min:%d Size:%d WH:%d WL:%d Params:%v",
		c.Name,
		c.MaximumTotalSizeGb,
		c.MinimumTotalSizeGb,
		c.DiskSizeGb,
		c.WatermarkHigh,
		c.WatermarkLow,
		c.Parameters)

}
