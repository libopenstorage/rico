/*
Package topology defines how to get information from the infrastructure
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
package topology

import (
	"fmt"
)

// Verify returns an error if the pool has any missing data
func (p *Pool) Verify() error {
	if p.SetSize == 0 {
		return fmt.Errorf("Size in pool cannot be zero")
	}
	if len(p.Class) == 0 {
		return fmt.Errorf("Pool class type cannot be empty")
	}
	return nil
}
