#!/bin/bash -x
# Copyright 2016 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


    MANAGED_PLUGIN_NAME="vsphere:latest"
    E2E_Tests="github.com/vmware/docker-volume-vsphere/tests/e2e"
    GO="go"
    SSH="ssh -i /root/.ssh/id_rsa -kTax -o StrictHostKeyChecking=no"

    get_vib_url() {
        echo "Get version $1"
        if [ $1 == "0.14" ]
        then
            VIB_URL="https://bintray.com/vmware/vDVS/download_file?file_path=VMWare_bootbank_esx-vmdkops-service_0.14.0577889-0.0.1.vib"
        elif [ $1 == "0.15" ]
        then
            VIB_URL="https://bintray.com/vmware/vDVS/download_file?file_path=VMWare_bootbank_esx-vmdkops-service_0.15.b93c186-0.0.1.vib"
        fi
    }

    VIB_URL=""
    get_vib_url $UPGRADE_FROM_VER
    FROM_VIB_URL=$VIB_URL
    echo "FROM_VIB_URL=$FROM_VIB_URL"

    get_vib_url $UPGRADE_TO_VER
    FROM_VIB_URL=$VIB_URL
    echo "TO_VIB_URL=$TO_VIB_URL"

    echo "Upgrade test: from ver $UPGRADE_FROM_VER to ver $UPGRADE_TO_VER"

	echo "Upgrade test step 1: deploy on $ESX with $FROM_VIB_URL"
	../misc/scripts/deploy-tools.sh deployesxForUpgrade $ESX $FROM_VIB_URL

	echo "Upgrade test step 2.1: remove plugin $MANAGED_PLUGIN_NAME on $VM1"
	../misc/scripts/deploy-tools.sh cleanvm $VM1 $MANAGED_PLUGIN_NAME

	echo "Upgrade test step 2.2: deploy plugin vmware/docker-volume-vsphere:$UPGRADE_FROM_VER on $VM1"
	../misc/scripts/deploy-tools.sh deployvm $VM1 vmware/docker-volume-vsphere:$UPGRADE_FROM_VER
	$SSH $VM1 "systemctl restart docker || service docker restart"

	echo "Upgrade test step 3: run pre-upgrade test"
	$GO test -v -timeout 30m -tags runpreupgrade $E2E_Tests

	echo "Upgrade test step 4: deploy on $ESX with $TO_VIB_URL"
	../misc/scripts/deploy-tools.sh deployesxForUpgrade  $ESX $TO_VIB_URL

	echo "Upgrade test step 5.1: remove plugin $MANAGED_PLUGIN_NAME on $VM1"
	../misc/scripts/deploy-tools.sh cleanvm $VM1 $MANAGED_PLUGIN_NAME

	echo "Upgrade test step 5.2: deploy plugin vmware/docker-volume-vsphere:$UPGRADE_TO_VER on $VM1"
	../misc/scripts/deploy-tools.sh deployvm $VM1 vmware/docker-volume-vsphere:$UPGRADE_TO_VER
	$SSH $VM1 "systemctl restart docker || service docker restart"

	echo "Upgrade test step 6: run pre-upgrade test"
	$GO test -v -timeout 30m -tags runpostupgrade $E2E_Tests