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
package main

import (
	"fmt"
	"strconv"
	"strings"

	fakecloud "github.com/libopenstorage/rico/pkg/cloudprovider/fake"
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/inframanager"
	fakestorage "github.com/libopenstorage/rico/pkg/storageprovider/fake"
	"github.com/libopenstorage/rico/pkg/topology"

	"github.com/abiosoft/ishell"
)

/*
ca name=small wh=60 wl=10 size=1 max=20 min=10

ca name=large wh=75 wl=25 size=250 max=10240 min=1024
*/

func main() {
	fc := fakecloud.New()
	fs := fakestorage.New(&topology.Topology{})
	class := config.Class{
		Name:               "gp2",
		WatermarkHigh:      75,
		WatermarkLow:       25,
		DiskSizeGb:         8,
		MaximumTotalSizeGb: 1024,
		MinimumTotalSizeGb: 32,
	}
	configuration := &config.Config{
		Classes: []config.Class{class},
	}
	im := inframanager.NewManager(configuration, fc, fs)

	// ishell
	shell := ishell.New()
	shell.Println("Rico Simulator")
	shell.SetPrompt("> ")

	// Node add
	shell.AddCmd(&ishell.Cmd{
		Name:    "node-add",
		Aliases: []string{"na"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("node-add <id>"))
				return
			}
			fs.NodeAdd(&topology.StorageNode{
				Name: c.Args[0],
				Metadata: topology.InstanceMetadata{
					ID: c.Args[0],
				},
			})
		},
		Help: "Add a node to the storage system",
	})

	// Utilization set
	shell.AddCmd(&ishell.Cmd{
		Name:    "utilization-set",
		Aliases: []string{"us"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(fmt.Errorf("utilization-set <class-name> <int>"))
				return
			}
			utilization, err := strconv.Atoi(c.Args[1])
			if err != nil {
				c.Err(err)
				return
			}

			found := false
			for _, class := range configuration.Classes {
				if class.Name == c.Args[0] {
					found = true
					fs.SetUtilization(&class, utilization)
					break
				}
			}
			if found {
				c.Println("OK")
			} else {
				c.Err(fmt.Errorf("class %s not found", c.Args[0]))
			}

		},
		Help: "Set utilization of a class across the cluster",
	})

	// Show topology
	shell.AddCmd(&ishell.Cmd{
		Name:    "topology",
		Aliases: []string{"t"},
		Func: func(c *ishell.Context) {
			t, _ := fs.GetTopology()
			c.Println("TOPOLOGY")
			c.Printf("Nodes: %d\n", len(t.Cluster.StorageNodes))
			c.Printf("Devices: %d\n", t.NumDevices())
			for _, class := range im.Config().Classes {
				c.Printf("C[%s|%dGi|%d] ",
					class.Name,
					t.TotalStorage(&class),
					t.Utilization(&class))
			}
			c.Println("")
			c.Println(t)
		},
		Help: "Show storage topoology",
	})

	// Reconcile
	shell.AddCmd(&ishell.Cmd{
		Name:    "reconcile",
		Aliases: []string{"r"},
		Func: func(c *ishell.Context) {
			if err := im.Reconcile(); err == nil {
				c.Println("OK")
			} else {
				c.Err(err)
			}
		},
		Help: "Reconcile once",
	})

	// List classes
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-list",
		Aliases: []string{"c", "classes"},
		Func: func(c *ishell.Context) {
			for _, class := range configuration.Classes {
				c.Printf("%v\n", class)
			}
		},
		Help: "List classes",
	})

	// Delete class
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-delete",
		Aliases: []string{"cd"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("Missing class name: class-delte <name>"))
				return
			}
			className := c.Args[0]

			// This should be part of config
			found := false
			index := 0
			for i, class := range configuration.Classes {
				if class.Name == className {
					found = true
					index = i
					break
				}
			}
			if !found {
				c.Err(fmt.Errorf("Class %s not found", className))
				return
			}

			// Delete class
			configuration.Classes[index] = configuration.Classes[len(configuration.Classes)-1]
			configuration.Classes = configuration.Classes[:len(configuration.Classes)-1]
			im.SetConfig(configuration)
			c.Println("OK")
		},
		Help: "Delete class",
	})

	// Add class
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-add",
		Aliases: []string{"ca"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 6 {
				c.Err(fmt.Errorf("Missing arguments: " +
					"ca name=<name> " +
					"wh=<watermark high> " +
					"wl=<watermark low> " +
					"size=<disk size Gi> " +
					"max=<total max size Gi> " +
					"min=<total min size Gi>"))
				return
			}
			newClass := config.Class{}
			for _, param := range c.Args {
				kv := strings.Split(strings.ToLower(param), "=")
				if len(kv) != 2 {
					c.Err(fmt.Errorf("Bad param: %s", param))
					return
				}
				switch kv[0] {
				case "name":
					newClass.Name = kv[1]
				case "wh":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						c.Err(err)
						return
					}
					newClass.WatermarkHigh = i
				case "wl":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						c.Err(err)
						return
					}
					newClass.WatermarkLow = i
				case "size":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.DiskSizeGb = i
				case "max":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.MaximumTotalSizeGb = i
				case "min":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.MinimumTotalSizeGb = i
				default:
					c.Err(fmt.Errorf("Unknown key: %s", kv[0]))
					return
				}
			}
			if len(newClass.Name) == 0 {
				c.Err(fmt.Errorf("Name missing: name=<name>"))
				return
			}
			if newClass.WatermarkHigh == 0 || newClass.WatermarkLow == 0 {
				c.Err(fmt.Errorf("Watermarks missing: wh=<int> wl=<int>"))
				return
			}
			if newClass.MinimumTotalSizeGb == 0 || newClass.MaximumTotalSizeGb == 0 {
				c.Err(fmt.Errorf("Max or min missing: max=<int> min=<int>"))
				return
			}
			if newClass.DiskSizeGb == 0 {
				c.Err(fmt.Errorf("Size missing: size=<int>"))
				return
			}
			configuration.Classes = append(configuration.Classes, newClass)
			im.SetConfig(configuration)
			c.Println("OK")
		},
		Help: "Add class",
	})

	// Run shell
	shell.Run()
	shell.Close()
}
