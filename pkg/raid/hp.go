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
	"encoding/json"
	"fmt"
)

type HP struct {
	execPath string
}

func NewHp(execPath string) *HP {
	return &HP{
		execPath: execPath,
	}
}

// GetControllersIDs - get number of controllers in the system
func (h *HP) GetControllersIDs() []string {
	inputData := GetCommandOutput(h.execPath, "ctrl", "all", "show")
	return GetRegexpAllSubmatch(inputData, "in Slot (.*?)[\\s]")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (h *HP) GetLogicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(h.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "ld", "all", "show")
	return GetRegexpAllSubmatch(inputData, "logicaldrive (.*?)[\\s]")
}

// GetArrayStatus get array status
func (h *HP) GetArrayStatus(controllerID, arrayId string) []byte {
	type ReturnData struct {
		ArrayID       string `json:"arrayId"`
		Status        string `json:"status"`
		InterfaceType string `json:"interfacetype"`
	}

	inputData := GetCommandOutput(h.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "array", fmt.Sprintf("%s", arrayId), "show", "detail")
	status := GetRegexpSubmatch(inputData, "[\\s]{2}Status: (.*)")
	interfaceType := GetRegexpSubmatch(inputData, "Interface Type: (.*)")

	data := ReturnData{
		Status:        TrimSpacesLeftAndRight(status),
		InterfaceType: TrimSpacesLeftAndRight(interfaceType),
		ArrayID:       fmt.Sprintf("%s", arrayId),
	}

	return append(MarshallJSON(data, 0), "\n"...)
}

// GetLDStatus - get logical drive status
func (h *HP) GetLDStatus(controllerID string, deviceID string) []byte {
	type ReturnData struct {
		Array    string `json:"array"`
		Status   string `json:"status"`
		Size     string `json:"size"`
		DiskName string `json:"diskName"`
		Raid     string `json:"raid"`
	}

	inputData := GetCommandOutput(h.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "ld", deviceID, "show", "detail")
	status := GetRegexpSubmatch(inputData, "Status *: (.*)")
	size := GetRegexpSubmatch(inputData, "Size *: (.*)")
	diskName := GetRegexpSubmatch(inputData, "Disk Name *: (.*)")
	arrayId := GetRegexpSubmatch(inputData, "[\\s]{2}Array (.*?)[\\s]")
	raid := GetRegexpSubmatch(inputData, "Fault Tolerance *: (.*)")

	data := ReturnData{
		Status:   TrimSpacesLeftAndRight(status),
		Size:     TrimSpacesLeftAndRight(size),
		DiskName: TrimSpacesLeftAndRight(diskName),
		Array:    TrimSpacesLeftAndRight(arrayId),
		Raid:     TrimSpacesLeftAndRight(raid),
	}

	return append(MarshallJSON(data, 0), "\n"...)
}

func (h *HP) GetVDisk(devName string) *RaidDisk {

	var (
		disk *RaidDisk
		data map[string]interface{}
	)

	controllersIDs := h.GetControllersIDs()
	for _, ctID := range controllersIDs {
		plogicalDrivesIDs := h.GetLogicalDrivesIDs(ctID)
		for _, ldID := range plogicalDrivesIDs {
			jld := h.GetLDStatus(ctID, ldID)
			if err := json.Unmarshal(jld, &data); err != nil {
				fmt.Printf("Error unmarshalling JSON data: %s\n", err)
			}
			if devName == data["diskName"].(string) {
				disk = &RaidDisk{
					// CT:        ctID,
					// ARR:       data["array"].(string),
					// LD:        ldID,
					DevPath: data["diskName"].(string),
					Raid:    data["raid"].(string),
					// Size:      data["size"].(uint64),
					MediaType: gettype(ctID, data["array"].(string), h)}
			}
		}
	}
	return disk
}

// gettype return array interface type
func gettype(ctID, arrayID string, h *HP) string {
	var data map[string]interface{}

	jar := h.GetArrayStatus(ctID, arrayID)
	if err := json.Unmarshal(jar, &data); err != nil {
		fmt.Printf("Error unmarshalling JSON data: %s\n", err)
	}
	intType := data["interfacetype"].(string)

	return intType
}
