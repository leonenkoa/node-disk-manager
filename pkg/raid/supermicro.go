/*
Copyright 2019 The OpenEBS Authors

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

package raid

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type SuperMicro struct {
	execPath string
}

type vDisk struct {
	ControllerID  int
	DiskID        string
	Raid          string
	DriveName     string
	DiskType      string
	InterfaceType string
}

func NewSuperMicro(execPath string) *SuperMicro {
	return &SuperMicro{
		execPath: execPath,
	}
}

// GetControllersIDs - get number of controllers in the system
func (s *SuperMicro) GetControllers() int {

	inputData := GetCommandOutput(s.execPath, "show", "J")
	out := gjson.Get(string(inputData), "Controllers.#.Response Data.Number of Controllers")

	controllers := out.Array()[0].Int()

	return int(controllers)
}

// GetVDs - get number of virtual disks in the system
func (s *SuperMicro) GetVDs() []string {

	inputData := GetCommandOutput(s.execPath, "/call/vall", "show", "J")
	out := gjson.Get(string(inputData), "Controllers.#.Response Data.Virtual Drives.#.DG/VD")

	var vds []string
	for _, vd := range out.Array() {
		for _, v := range vd.Array() {
			vds = append(vds, v.String())
		}
	}

	return vds
}

// GetVDisks - get info about all disks,types and i.e in the system
func (s *SuperMicro) GetVDisk(devName string) *RaidDisk {
	ctrls := s.GetControllers()
	vds := s.GetVDs()

	ctrl := 0
	var disk *RaidDisk
	for ctrl < ctrls {
		ctrl++
		for _, vd := range vds {
			v := strings.Split(vd, "/")
			inputData := GetCommandOutput(s.execPath, fmt.Sprintf("/c%v/v%s", ctrl-1, v[1]), "show", "all", "J")
			if devName == gjson.Get(string(inputData), "Controllers.0.Response Data.VD* Properties.OS Drive Name").String() {
				disk = &RaidDisk{
					// ControllerID:  ctrl - 1,
					// DiskID:        vd,
					Raid:      gjson.Get(string(inputData), "Controllers.0.Response Data./c*v*.#.TYPE").Array()[0].String(),
					DevPath:   gjson.Get(string(inputData), "Controllers.0.Response Data.VD* Properties.OS Drive Name").String(),
					MediaType: gjson.Get(string(inputData), "Controllers.0.Response Data.PDs for VD *.#.Med").Array()[0].String(),
					// InterfaceType: gjson.Get(string(inputData), "Controllers.0.Response Data.PDs for VD *.#.Intf").Array()[0].String(),
				}
			}
		}
	}

	return disk
}
