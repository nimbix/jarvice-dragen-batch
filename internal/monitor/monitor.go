/*
Copyright (c) 2023, Nimbix, Inc.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation are
those of the authors and should not be interpreted as representing official
policies, either expressed or implied, of Nimbix, Inc.
*/

package monitor

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Meter interface {
	Init() bool
	Running() bool
	ExitSuccess() bool
	Cleanup()
	Output()
}

func StartMonitor(meter Meter) error {
	catchSignal := make(chan os.Signal, 1)
	signal.Notify(catchSignal, syscall.SIGINT, syscall.SIGTERM)

	interval, err := strconv.ParseInt(os.Getenv("JARVICE_POLL_INTERVAL"), 10, 64)
	if err != nil {
		interval = 5
	}

	if !meter.Init() {
		meter.Cleanup()
		return errors.New("monitor init failed")
	}

	timer := time.NewTicker(time.Duration(interval) * time.Second)
	defer timer.Stop()
Poll:
	for {
		select {
		case <-catchSignal:
			break Poll
		case <-timer.C:
			if !meter.Running() {
				break Poll
			}
			meter.Output()
		}
	}

	success := meter.ExitSuccess()
	meter.Cleanup()
	if success {
		return nil
	} else {
		return errors.New("job processing failed")
	}
}
