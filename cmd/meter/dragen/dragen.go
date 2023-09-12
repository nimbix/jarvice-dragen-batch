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

package dragen

import (
	"errors"

	"jarvice.io/dragen/internal/google"
	"jarvice.io/dragen/internal/jobs"
	"jarvice.io/dragen/internal/logger"
)

type DragenMeter struct {
	job         *jobs.JarviceJob
	vm          *google.GoogleCompute
	ServiceName string
}

func NewDragenMeter(apiHost, username, apikey, jobId, serviceName string) (*DragenMeter, error) {
	job := jobs.NewJarviceJob(apiHost, username, apikey, jobId)
	if !job.CheckAuth() {
		return nil, errors.New("invalid JARVICE configuration")
	}
	if vm, err := google.NewGoogleCompute(); err != nil {
		return nil, errors.New("unable to create Google Cloud client")
	} else {
		return &DragenMeter{
			job:         job,
			vm:          vm,
			ServiceName: serviceName,
		}, nil
	}
}

func (meter DragenMeter) Init() bool {
	return google.CheckDragenLicense()
}

func (meter DragenMeter) Running() bool {
	return meter.job.Running() && meter.vm.InstanceExist(meter.ServiceName)
}

func (meter DragenMeter) Cleanup() {
	meter.job.Terminate()
	if meter.vm != nil {
		name := meter.vm.GetName()
		meter.vm.DeleteReservationWait(name, false)
		meter.vm.DeleteTemplateWait(name, false)
		meter.vm.DeleteHost()
	} else {
		logger.Elogger.Error("cannot remove Google Cloud objects. Please verify removal of the template, reservation, and vm for this job")
	}
}

func (meter DragenMeter) Output() {
}

func (meter DragenMeter) ExitSuccess() bool {
	return meter.job.ExitSuccess()
}
