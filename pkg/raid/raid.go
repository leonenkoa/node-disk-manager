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

import "C"
import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/zcalusic/sysinfo"
)

// Identifier (devPath such as /dev/sda,etc) is an identifier for seachest probe
type Identifier struct {
	DevPath string
}

type (
	RaidDisk struct {
		DevPath   string `json:"devPath"`
		Raid      string `json:"raid"`
		MediaType string `json:"mediatype"`
	}
	RaidDisks struct {
		RaidDisksList []RaidDisk `json:"data"`
	}
)

func (I *Identifier) GetVendor() string {

	current, err := user.Current()
	if err != nil {
		panic(err)
	}

	if current.Uid != "0" {
		fmt.Println("requires superuser privilege")
	}

	var (
		si sysinfo.SysInfo
	)

	si.GetSysInfo()

	vendor := strings.ToLower(si.Board.Vendor)

	TrimSpacesLeftAndRight(vendor)

	return vendor
}

func (I *Identifier) RaidDiskInfo(vendor, devPath string) *RaidDisk {
	type Config struct {
		Vendors interface{} `json:"vendors"`
	}

	var configFile = []byte(`
{
    "vendors": {
        "hp": "ssacli",
		"dell inc.": "/opt/dell/srvadmin/sbin/omcliproxy",
		"supermicro": "/opt/MegaRAID/storcli/storcli64"
    }
}
`)

	var (
		configJSON Config
		vendors    []string
		toolVendor string
		toolBinary string
	)

	if err := json.Unmarshal(configFile, &configJSON); err != nil {
		fmt.Printf("Error unmarshalling JSON data: %s\n", err)
		os.Exit(1)
	}

	if configJSON.Vendors == nil {
		fmt.Println("Failed to get vendors from config file.")
		os.Exit(1)
	}

	for v := range configJSON.Vendors.(map[string]interface{}) {
		vendors = append(vendors, v)
	}

	toolVendor = vendor

	for i, v := range vendors {
		if v != toolVendor {
			if i == len(vendors)-1 {
				fmt.Printf("Vendors must be one of '%s', got '%s'.\n", strings.Join(vendors, " | "), toolVendor)
			}
			continue
		}
		break
	}

	toolBinary = configJSON.Vendors.(map[string]interface{})[toolVendor].(string)

	switch toolVendor {
	case "hp":
		h := NewHp(toolBinary)
		res := h.GetVDisk(devPath)
		return res
	case "dell inc.":
		d := NewDell(toolBinary)
		res := d.GetVDisk(devPath)
		return res
	case "supermicro":
		s := NewSuperMicro(toolBinary)
		res := s.GetVDisk(devPath)
		return res
	}
	return nil
}
