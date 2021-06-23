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
	"errors"
	"strings"
)

const (
	ecsAgentEndpoint         = "http://%s:51678/v1/metadata"
	ecsAgentTaskInfoEndpoint = "http://%s:51678/v1/tasks"
	taskStatusRunning        = "RUNNING"

	// infinity magic number for cgroup: https://unix.stackexchange.com/questions/420906/what-is-the-value-for-the-cgroups-limit-in-bytes-if-the-memory-is-not-restricte
	kernelMagicCodeNotSet = int64(9223372036854771712)

	ecsInstanceMountConfigPath = "/proc/self/mountinfo"
)

// There are two formats of ContainerInstance ARN (https://docs.aws.amazon.com/AmazonECS/latest/userguide/ecs-account-settings.html#ecs-resource-ids)
// arn:aws:ecs:region:aws_account_id:container-instance/container-instance-id
// arn:aws:ecs:region:aws_account_id:container-instance/cluster-name/container-instance-id
// This function will return "container-instance-id" for both ARN format

func GetContainerInstanceIdFromArn(arn string) (containerInstanceId string, err error) {
	// When splitting the ARN with ":", the 6th segments could be either:
	// container-instance/47c0ab6e-2c2c-475e-9c30-b878fa7a8c3d or
	// container-instance/cluster-name/47c0ab6e-2c2c-475e-9c30-b878fa7a8c3d
	err = nil
	if splitedList := strings.Split(arn, ":"); len(splitedList) >= 6 {
		// Further splitting tmpResult with "/", it could be splitted into either 2 or 3
		// Characters of "cluster-name" is only allowed to be letters, numbers and hyphens
		tmpResult := strings.Split(splitedList[5], "/")
		if len(tmpResult) == 2 {
			containerInstanceId = tmpResult[1]
			return
		} else if len(tmpResult) == 3 {
			containerInstanceId = tmpResult[2]
			return
		}
	}
	err = errors.New("Can't get ecs container instance id from ContainerInstance arn: " + arn)
	return

}
