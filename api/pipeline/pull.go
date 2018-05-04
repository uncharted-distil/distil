package pipeline

import (
	"io"
	"time"

	"github.com/pkg/errors"
)

const (
	pullTimeout = 10 * time.Second
	pullMax     = 10
)

type pullFunc func() error

func pullFromAPI(maxPulls int, timeout time.Duration, pull pullFunc) error {

	recvChan := make(chan error)

	count := 0
	for {

		// pull
		go func() {
			err := pull()
			recvChan <- err
		}()

		// set timeout timer
		timer := time.NewTimer(timeout)

		select {
		case err := <-recvChan:
			timer.Stop()
			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
			count++
			if count > maxPulls {
				return nil
			}

		case <-timer.C:
			// timeout
			return errors.Errorf("solution request has timed out")
		}

	}
}
