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

package batch

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"jarvice.io/dragen/cmd/service/jarvice"
	"jarvice.io/dragen/config"
	"jarvice.io/dragen/internal/google"
	"jarvice.io/dragen/internal/jobs"
	"jarvice.io/dragen/internal/logger"
)

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

const vmBaseName = "dragen"

func cleanup(label string) {
	if vm, err := google.NewGoogleCompute(); err == nil {
		logger.Ologger.Warn("Google Compute Engine objects being removed for " + label)
		vm.DeleteInstanceWait(label, false)
		vm.DeleteReservation(label)
		vm.DeleteTemplate(label)
	} else {
		logger.Elogger.Error("unable to remove Google Compute Engine object (template, reservation, and vm) for " + label)
	}
}

type DragenBatch struct {
	label                              string
	serviceAccount                     string
	vm                                 *google.GoogleCompute
	job                                *jobs.JarviceJob
	b64Args                            string
	app                                string
	apiHost, username, apikey, machine string
	priority                           string
}

func NewDragenBatch(apiHost, username, apikey, app, machine,
	s3AccessKey, s3SecretKey, illuminaLic string,
	serviceAccount, priority string, args ...string) (*DragenBatch, error) {
	if len(args) < 1 {
		return nil, errors.New("missing Dragen arguments")
	}

	dragenBatch := DragenBatch{}
	dragenBatch.label = vmBaseName + "-" + randomString(12)
	dargs := []string{}
	if len(s3AccessKey) < 1 {
		return nil, errors.New("missing --s3-access-key")
	} else {
		dargs = append(dargs, "--s3-access-key")
		dargs = append(dargs, s3AccessKey)
	}
	if len(s3SecretKey) < 1 {
		return nil, errors.New("missing --s3-secret-key")
	} else {
		dargs = append(dargs, "--s3-secret-key")
		dargs = append(dargs, s3SecretKey)
	}
	dargs = append(dargs, args...)
	if len(illuminaLic) > 0 {
		dargs = append(dargs, "--lic-server")
		dargs = append(dargs, illuminaLic)
	}

	dragenBatch.b64Args = base64.RawStdEncoding.EncodeToString([]byte(strings.Join(dargs, " ")))
	if vm, err := google.NewGoogleCompute(); err != nil {
		return nil, err
	} else {
		dragenBatch.vm = vm
	}

	dragenBatch.serviceAccount = serviceAccount
	dragenBatch.apiHost = apiHost
	dragenBatch.username = username
	dragenBatch.apikey = apikey
	dragenBatch.app = app
	dragenBatch.machine = machine
	dragenBatch.job = &jobs.JarviceJob{}
	dragenBatch.priority = priority
	return &dragenBatch, nil
}

func (b DragenBatch) Init() bool {

	if err := b.vm.CreateInstanceTemplates(b.label, b.serviceAccount,
		config.MeterContainer+":"+config.Version,
		"/usr/local/bin/entrypoint",
		"--api-host", b.apiHost,
		"--username", b.username,
		"--apikey", b.apikey,
		"--job-id", "TEMP_JOB_ID",
		"--service-name", b.vm.GetName(),
	); err != nil {
		logger.Ologger.Warn(err.Error())
		cleanup(b.label)
		return false
	}

	if err := b.vm.CreateReservation(b.label, b.label); err != nil {
		logger.Ologger.Warn(err.Error())
		cleanup(b.label)
		return false
	}
	if number, err := jarvice.SubmitJarviceJob(b.app, b.machine,
		b.vm.GetId(), b.vm.GetProject(), b.vm.GetZone(),
		b.apiHost, b.username, b.apikey, b.priority, b.b64Args); err != nil {
		logger.Ologger.Warn(err.Error())
		cleanup(b.label)
		return false
	} else {
		*b.job = *(jobs.NewJarviceJob(b.apiHost, b.username, b.apikey, number))
	}
	for {
		running, err := b.job.RunningWithError()
		if err != nil {
			logger.Ologger.Warn(err.Error())
			cleanup(b.label)
			return false
		}
		if running {
			break
		}
		// wait
		time.Sleep(15 * time.Second)
	}
	if err := b.vm.CreateInstanceWithJobId(b.label, b.label,
		b.label, b.job.Number, b.job.GoogleShutdownScript()); err != nil {
		logger.Ologger.Warn(err.Error())
		cleanup(b.label)

		return false
	}
	logger.Ologger.Info("Batch processing starting")

	return true
}

func (b DragenBatch) Running() bool {
	// TODO: add check for MP VM
	return b.job.Running()
}

func (b DragenBatch) Cleanup() {
	logger.Ologger.Info("running cleanup")
	if b.job != nil {
		b.job.Terminate()
	}
	if b.vm != nil {
		b.vm.DeleteInstanceWait(b.label, false)
		b.vm.DeleteReservation(b.label)
		b.vm.DeleteTemplate(b.label)
	}
}

func (b DragenBatch) Output() {
	b.job.GetJobOutput()
}

func (b DragenBatch) ExitSuccess() bool {
	if b.job != nil {
		return b.job.ExitSuccess()
	}
	return false
}
