// Copyright 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This util is holding misc small functions for operations to be done using admincli on esx

package admincli

import (
	"log"
	"strings"

	"github.com/vmware/docker-volume-vsphere/tests/constants/admincli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/ssh"
)

// UpdateVolumeAccess update the volume access as per params
func UpdateVolumeAccess(ip, volName, vmgroup, access string) (string, error) {
	log.Printf("Updating access to [%s] for volume [%s] ", access, volName)
	return ssh.InvokeCommand(ip, admincli.SetVolumeAccess+" --vmgroup="+vmgroup+
		" --volume="+volName+" --options=\"access="+access+"\"")
}

// GetAllVolumeProperties returns a map of all the volumes from ESX and their corresponding
// properties - capacity, attached-to-vm and disk-format.
func GetAllVolumeProperties(hostName string) map[string][]string {
	log.Printf("Getting size, disk-format and attached-to-vm for all volumes from ESX [%s].", hostName)
	cmd := admincli.ListVolumes + "-c volume,capacity,disk-format,attached-to 2>/dev/null | awk -v OFS='\t' '{print $1, $2, $3, $4}' | sed '1,2d'  "
	out, _ := ssh.InvokeCommand(hostName, cmd)
	admincliValues := strings.Fields(out)
	adminCliMap := make(map[string][]string)
	for i := 0; i < len(admincliValues); {
		adminCliMap[admincliValues[i]] = []string{admincliValues[i+1], admincliValues[i+2], admincliValues[i+3]}
		i = i + 4
	}
	return adminCliMap
}

// GetVolumeProperties returns an array of properties of a particular volume from ESX.
// Properties returned - capacity, attached-to-vm and disk-format field
func GetVolumeProperties(volumeName, hostName string) []string {
	log.Printf("Getting size, disk-format and attached-to-vm for volume [%s] from ESX [%s]. \n", volumeName, hostName)
	cmd := admincli.ListVolumes + "-c volume,capacity,disk-format,attached-to 2>/dev/null | grep " + volumeName + " | awk -v OFS='\t' '{print $2, $3, $4}' "
	out, _ := ssh.InvokeCommand(hostName, cmd)
	return strings.Fields(out)
}
