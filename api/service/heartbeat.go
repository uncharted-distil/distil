//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package service

import (
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
)

// Heartbeat of a service.
type Heartbeat func() bool

// IsUp checks if the service is available.
func IsUp(test Heartbeat) bool {
	return test()
}

// WaitForService waits for the service to become available.
func WaitForService(serviceName string, config *env.Config, test Heartbeat) error {
	up := false
	i := 0
	retryCount := config.ServiceRetryCount
	for ; i < retryCount && !up; i++ {
		log.Infof("Waiting for service '%s' (attempt %d)", serviceName, i+1)
		if IsUp(test) {
			up = true
		} else {
			time.Sleep(10 * time.Second)
		}
	}

	if i == retryCount {
		return errors.Errorf("unable to connect to service '%s' after %d attempts", serviceName, retryCount)
	}

	return nil
}
