package service

import (
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"

	"github.com/unchartedsoftware/distil/api/env"
)

type Heartbeat func() bool

func ServiceIsUp(test Heartbeat) bool {
	return test()
}

func WaitForService(serviceName string, config *env.Config, test Heartbeat) error {
	up := false
	i := 0
	retryCount := config.ServiceRetryCount
	for ; i < retryCount && !up; i++ {
		log.Infof("Waiting for service '%s' (attempt %d)", serviceName, i+1)
		if ServiceIsUp(test) {
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
