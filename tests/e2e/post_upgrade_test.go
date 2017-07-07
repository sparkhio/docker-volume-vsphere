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

// This test suite includes test cases to verify basic functionality
// in most common configurations

// +build runpostupgrade

package e2e

import (
	"github.com/vmware/docker-volume-vsphere/tests/utils/dockercli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/inputparams"
	"github.com/vmware/docker-volume-vsphere/tests/utils/misc"
	"github.com/vmware/docker-volume-vsphere/tests/utils/verification"
	. "gopkg.in/check.v1"
)

type PostUpgradeTestSuite struct {
	config         *inputparams.TestConfig
	esx            string
	vm1            string
	vm2            string
	vm1Name        string
	vm2Name        string
	volName1       string
	volName2       string
	containerName1 string
	containerName2 string
}

func (s *PostUpgradeTestSuite) SetUpSuite(c *C) {
	s.config = inputparams.GetTestConfig()
	if s.config == nil {
		c.Skip("Unable to retrieve test config, skipping basic tests")
	}

	s.esx = s.config.EsxHost
	s.vm1 = s.config.DockerHosts[0]
	s.vm2 = s.config.DockerHosts[1]
	s.vm1Name = s.config.DockerHostNames[0]
	s.vm2Name = s.config.DockerHostNames[1]
}

func (s *PostUpgradeTestSuite) SetUpTest(c *C) {
	s.volName1 = "preUpgradeTestVol"
	s.volName2 = "postUpgradeTestVol"
	s.containerName1 = inputparams.GetUniqueContainerName(c.TestName())
	s.containerName2 = inputparams.GetUniqueContainerName(c.TestName())
}

var _ = Suite(&PostUpgradeTestSuite{})

// Test steps:
// 1. Verify volume created in pre-upgrade test is available
// 2. Verify the volume is detached
// 3. Attach the volume from a new conatainer
// 4. Verify volume status is attached
// 5. Read data from that volume to make sure data written in pre-upgrade test remains same
// 5. Write more data to that volume
// 6. Read the data from the volume to make sure it matches the data written in previous step
// 7. Stop the container and verify volume status is detached
// 8. Remove the container
// 9. Remove the volume created in pre-upgrade test
// 10. Create a new volume
// 11. Verify the status is detached
// 12. Attach the new volume from a new container
// 13. Verify the volume status is attached
// 14. Write some data to the new volume
// 15. Read the data from the new volume to make it matches the data written in previous step
// 16. Stop the container
// 17. Remove the container
// 18. Remove the new volume
func (s *PostUpgradeTestSuite) TestVolumeLifecycle(c *C) {
	const (
		testData  = "Hello World!"
		testFile  = "hello.txt"
		testData1 = "Hello World Again!"
		testFile1 = "hello1.txt"
		volPath   = "/vol"
	)
	misc.LogTestStart(c.TestName())

	accessible := verification.CheckVolumeAvailability(s.vm1, s.volName1)
	c.Assert(accessible, Equals, true, Commentf("Volume %s is not available on docker host %s",
		s.volName1, s.vm1))

	status := verification.VerifyDetachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not detached", s.volName1))

	out, err := dockercli.AttachVolume(s.vm1, s.volName1, s.containerName1)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyAttachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not attached", s.volName1))

	out, err = dockercli.ReadFromContainer(s.vm1, s.containerName1, volPath, testFile)
	c.Assert(err, IsNil, Commentf(out))
	c.Assert(out, Equals, testData)

	out, err = dockercli.WriteToContainer(s.vm1, s.containerName1, volPath, testFile1, testData1)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.ReadFromContainer(s.vm1, s.containerName1, volPath, testFile1)
	c.Assert(err, IsNil, Commentf(out))
	c.Assert(out, Equals, testData1)

	out, err = dockercli.StopContainer(s.vm1, s.containerName1)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyDetachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not detached", s.volName1))

	out, err = dockercli.RemoveContainer(s.vm1, s.containerName1)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.DeleteVolume(s.vm1, s.volName1)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.CreateVolume(s.vm1, s.volName2)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyDetachedStatus(s.volName2, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not detached", s.volName2))

	out, err = dockercli.AttachVolume(s.vm1, s.volName2, s.containerName2)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyAttachedStatus(s.volName2, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not attached", s.volName2))

	out, err = dockercli.WriteToContainer(s.vm1, s.containerName2, volPath, testFile, testData)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.ReadFromContainer(s.vm1, s.containerName2, volPath, testFile)
	c.Assert(err, IsNil, Commentf(out))
	c.Assert(out, Equals, testData)

	out, err = dockercli.StopContainer(s.vm1, s.containerName2)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.RemoveContainer(s.vm1, s.containerName2)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.DeleteVolume(s.vm1, s.volName2)
	c.Assert(err, IsNil, Commentf(out))

}
