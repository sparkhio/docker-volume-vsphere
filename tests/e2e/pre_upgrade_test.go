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

// +build runpreupgrade

package e2e

import (
	"github.com/vmware/docker-volume-vsphere/tests/utils/dockercli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/inputparams"
	"github.com/vmware/docker-volume-vsphere/tests/utils/misc"
	"github.com/vmware/docker-volume-vsphere/tests/utils/verification"
	. "gopkg.in/check.v1"
)

type PreUpgradeTestSuite struct {
	config        *inputparams.TestConfig
	esx           string
	vm1           string
	vm2           string
	vm1Name       string
	vm2Name       string
	volName1      string
	volName2      string
	containerName string
}

func (s *PreUpgradeTestSuite) SetUpSuite(c *C) {
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

func (s *PreUpgradeTestSuite) SetUpTest(c *C) {
	s.volName1 = "preUpgradeTestVol"
	s.containerName = inputparams.GetUniqueContainerName(c.TestName())
}

var _ = Suite(&PreUpgradeTestSuite{})

// Test steps:
// 1. Create a volume
// 2. Verify the volume is detached
// 3. Attach the volume
// 4. Verify volume status is attached
// 5. Write some data to that volume
// 6. Read the data from the volume to make sure it matched the data written previously
// 7. Stop the container and verify volume status is detached
// 8. Start container and read from container again to make sure data remains same
// 9. Verify volume status is attached
// 10. Remove the container
func (s *PreUpgradeTestSuite) TestVolumeLifecycle(c *C) {
	const (
		testData = "Hello World!"
		testFile = "hello.txt"
		volPath  = "/vol"
	)
	misc.LogTestStart(c.TestName())

	out, err := dockercli.CreateVolume(s.vm1, s.volName1)
	c.Assert(err, IsNil, Commentf(out))

	status := verification.VerifyDetachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not detached", s.volName1))

	out, err = dockercli.AttachVolume(s.vm1, s.volName1, s.containerName)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyAttachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not attached", s.volName1))

	out, err = dockercli.WriteToContainer(s.vm1, s.containerName, volPath, testFile, testData)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.ReadFromContainer(s.vm1, s.containerName, volPath, testFile)
	c.Assert(err, IsNil, Commentf(out))
	c.Assert(out, Equals, testData)

	out, err = dockercli.StopContainer(s.vm1, s.containerName)
	c.Assert(err, IsNil, Commentf(out))

	status = verification.VerifyDetachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not detached", s.volName1))

	out, err = dockercli.StartContainer(s.vm1, s.containerName)
	c.Assert(err, IsNil, Commentf(out))

	out, err = dockercli.ReadFromContainer(s.vm1, s.containerName, volPath, testFile)
	c.Assert(err, IsNil, Commentf(out))
	c.Assert(out, Equals, testData)

	status = verification.VerifyAttachedStatus(s.volName1, s.vm1, s.esx)
	c.Assert(status, Equals, true, Commentf("Volume %s is not attached", s.volName1))

	out, err = dockercli.RemoveContainer(s.vm1, s.containerName)
	c.Assert(err, IsNil, Commentf(out))

}
