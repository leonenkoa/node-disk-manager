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

	"github.com/leonenkoa/gomreport"
)

type Dell struct {
	execPath string
}

func NewDell(execPath string) *Dell {
	return &Dell{
		execPath: execPath,
	}
}

// GetControllersIDs - get number of controllers in the system
func (d *Dell) GetVDisk(devName string) *RaidDisk {

	config := &omreport.Config{OMCLIProxyPath: d.execPath, EnhancedSecurityMode: false}

	om, err := omreport.NewOMReporter(config)

	if err != nil {
		fmt.Println(err)
	}

	vdk, err := om.StorageVDisk()

	if err != nil {
		fmt.Println(err)
	}

	var disk *RaidDisk
	for _, v := range vdk.VDisks {
		if v.DeviceName == devName {
			disk = &RaidDisk{
				DevPath: v.DeviceName,
				// BusProtocol: vd.BusProtocol.String(),
				Raid: v.Layout.String(),
				// Size:      uint64(v.Size),
				MediaType: v.MediaType.String(),
			}
		}
	}
	return disk
}
