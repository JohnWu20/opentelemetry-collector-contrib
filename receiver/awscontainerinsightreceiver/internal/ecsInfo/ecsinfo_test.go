// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ecsinfo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	httpClient "github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/containerinsight/httpclient"
)

type MochhostInfo struct{}

func (hi *MochhostInfo) GetInstanceIp() string {
	return "host-ip-address"
}
func (hi *MochhostInfo) GetClusterName() string {
	return ""
}
func (hi *MochhostInfo) GetinstanceIpReadyC() chan bool {
	readyC := make(chan bool)
	close(readyC)
	return readyC
}

type MockInstanceInfo struct {
	clusterName string
	instanceID  string
}

func (ii *MockInstanceInfo) GetClusterName() string {
	return ii.clusterName
}
func (ii *MockInstanceInfo) GetContainerInstanceId() string {
	return ii.instanceID
}

type MockTaskInfo struct {
	tasks            []ECSTask
	runningTaskCount int64
}

func (ii *MockTaskInfo) getRunningTaskCount() int64 {
	return ii.runningTaskCount
}
func (ii *MockTaskInfo) getRunningTasksInfo() []ECSTask {

	return ii.tasks
}

type MockCgroupScanner struct {
	cpuReserved int64
	memReserved int64
}

func (c *MockCgroupScanner) getCpuReserved() int64 {
	return c.memReserved
}

func (c *MockCgroupScanner) getMemReserved() int64 {
	return c.memReserved
}

func (c *MockCgroupScanner) getCPUReservedInTask(taskID string, clusterName string) int64 {
	return int64(10)
}

func (c *MockCgroupScanner) getMEMReservedInTask(taskID string, clusterName string, containers []ECSContainer) int64 {
	return int64(512)
}

func TestNewECSInfo(t *testing.T) {
	// test the case when containerInstanceInfor fails to initialize
	containerInstanceInfoCreatorOpt := func(ei *EcsInfo) {
		ei.containerInstanceInfoCreator = func(context.Context, hostIPProvider, time.Duration, *zap.Logger,
			httpClient.HttpClientProvider, chan bool, ...ecsInstanceInfoOption) containerInstanceInfoProvider {
			return &MockInstanceInfo{
				clusterName: "Cluster-name",
				instanceID:  "instance-id",
			}
		}
	}

	taskinfoCreatorOpt := func(ei *EcsInfo) {
		ei.ecsTaskInfoCreator = func(context.Context, hostIPProvider, time.Duration, *zap.Logger, httpClient.HttpClientProvider,
			chan bool, ...taskInfoOption) ecsTaskInfoProvider {
			tasks := []ECSTask{}
			return &MockTaskInfo{
				tasks:            tasks,
				runningTaskCount: int64(2),
			}
		}
	}

	cgroupScannerCreatorOpt := func(ei *EcsInfo) {
		ei.cgroupScannerCreator = func(context.Context, *zap.Logger, ecsTaskInfoProvider, containerInstanceInfoProvider,
			time.Duration) cgroupScannerProvider {
			return &MockCgroupScanner{
				cpuReserved: int64(20),
				memReserved: int64(1024),
			}
		}
	}
	hostIPProvider := &MochhostInfo{}

	ecsinfo, _ := NewECSInfo(time.Minute, hostIPProvider, zap.NewNop(), containerInstanceInfoCreatorOpt, taskinfoCreatorOpt, cgroupScannerCreatorOpt)

	assert.NotNil(t, ecsinfo)

	time.Sleep(time.Second * 1)

	assert.Equal(t, "instance-id", ecsinfo.GetContainerInstanceId())
	assert.Equal(t, "Cluster-name", ecsinfo.GetClusterName())
	assert.Equal(t, int64(2), ecsinfo.GetRunningTaskCount())

	assert.Equal(t, int64(0), ecsinfo.GetCpuReserved())
	assert.Equal(t, int64(0), ecsinfo.GetMemReserved())

	close(ecsinfo.isTaskInfoReadyC)
	close(ecsinfo.isContainerInfoReadyC)

	time.Sleep(time.Second * 1)
	assert.Equal(t, int64(1024), ecsinfo.GetCpuReserved())
	assert.Equal(t, int64(1024), ecsinfo.GetMemReserved())

}
