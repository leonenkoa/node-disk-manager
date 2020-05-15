/*
Copyright 2018 OpenEBS Authors.

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

package probe

import (
	"k8s.io/klog"

	"github.com/leonenkoa/node-disk-manager/pkg/raid"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// raid contains required variables for populating diskInfo
type raidProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller     *controller.Controller
	RaidIdentifier *raid.Identifier
}

const (
	raidConfigKey     = "raid-probe"
	raidProbePriority = 6
)

var (
	raidProbeName  = "raid probe"
	raidProbeState = defaultEnabled
)

// init is used to get a controller object and then register itself
var raidProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", raidProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == raidConfigKey {
				raidProbeName = probeConfig.Name
				raidProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   raidProbePriority,
		name:       raidProbeName,
		state:      raidProbeState,
		pi:         &raidProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the probe (seachest probe in this case)
	newRegisterProbe.register()
}

// newSeachestProbe returns seachestProbe struct which helps populate diskInfo struct
// with the basic disk details such as logical size, firmware revision, etc
func newRaidProbe(devPath string) *raidProbe {
	raidIdentifier := &raid.Identifier{
		DevPath: devPath,
	}
	raidProbe := &raidProbe{
		RaidIdentifier: raidIdentifier,
	}
	return raidProbe
}

// Start is mainly used for one time activities such as monitoring.
// It is a part of probe interface but here we does not require to perform
// such activities, hence empty implementation
func (rp *raidProbe) Start() {}

// fillDiskDetails fills details in diskInfo struct using information it gets from probe
func (rp *raidProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	if blockDevice.DevPath == "" {
		klog.Error("raidIdentifier is found empty, raid probe will not fill disk details.")
		return
	}

	raidProbe := newRaidProbe(blockDevice.DevPath)

	vendor := raidProbe.RaidIdentifier.GetVendor()

	driveInfo := raidProbe.RaidIdentifier.RaidDiskInfo(vendor, blockDevice.DevPath)

	if blockDevice.DeviceAttributes.DriveType == "" {
		blockDevice.DeviceAttributes.DriveType = driveInfo.MediaType
		klog.V(4).Infof("Disk: %s DriveType:%s filled by raid probe.", blockDevice.DevPath, blockDevice.DeviceAttributes.DriveType)
	}

	if blockDevice.DeviceAttributes.DeviceType == "" {
		blockDevice.DeviceAttributes.DriveType = driveInfo.Raid
		klog.V(4).Infof("Disk: %s DriveType:%s filled by raid probe.", blockDevice.DevPath, blockDevice.DeviceAttributes.DeviceType)
	}
}
