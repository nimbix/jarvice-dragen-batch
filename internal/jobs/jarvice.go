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

package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"jarvice.io/dragen/internal/logger"
)

const tailLength = 100
const compareWindow = 25

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type JobStatus struct {
	Status string `json:"job_status"`
}

type JobStatusList map[string]JobStatus

type JarviceJob struct {
	apiHost, Number string
	values          url.Values
	lastTail        []string
}

func NewJarviceJob(apiHost, username, apikey, number string) *JarviceJob {
	return &JarviceJob{
		apiHost: strings.TrimSuffix(apiHost, "/"),
		Number:  number,
		values: url.Values{
			"username": {username},
			"apikey":   {apikey},
			"number":   {number},
		},
		lastTail: make([]string, tailLength),
	}
}

func (job JarviceJob) CheckAuth() bool {

	resp, err := http.PostForm(job.apiHost+"/jarvice/machines", job.values)
	if err != nil {
		logger.Ologger.Warn(err.Error())
		return false
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

func (job JarviceJob) Running() bool {
	if ret, err := job.RunningWithError(); err != nil {
		logger.Ologger.Warn(err.Error())
		return false
	} else {
		return ret
	}
}

func (job JarviceJob) RunningWithError() (bool, error) {

	resp, err := http.PostForm(job.apiHost+"/jarvice/status", job.values)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("Cannot find JARVICE job " + job.Number)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	jobStatusList := JobStatusList{}
	json.Unmarshal(body, &jobStatusList)
	if jobStatusList[job.Number].Status == "PROCESSING STARTING" {
		return true, nil
	} else if jobStatusList[job.Number].Status == "SUBMITTED" {
		return false, nil
	} else {
		return false, errors.New("JARVICE job not found")
	}
}

func (job JarviceJob) ExitSuccess() bool {
	if ret, err := job.ExitSuccessWithError(); err != nil {
		logger.Ologger.Warn(err.Error())
		return false
	} else {
		return ret
	}
}

func (job JarviceJob) ExitSuccessWithError() (bool, error) {

	resp, err := http.PostForm(job.apiHost+"/jarvice/status", job.values)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("Cannot find JARVICE job " + job.Number)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	jobStatusList := JobStatusList{}
	json.Unmarshal(body, &jobStatusList)
	if jobStatusList[job.Number].Status == "COMPLETED" {
		return true, nil
	} else {
		return false, errors.New("JARVICE job failed")
	}
}

func (job JarviceJob) GetJobOutput() {
	defer func() {
		recover()
	}()
	resp, err := http.PostForm(job.apiHost+"/jarvice/tail", job.values)
	if err != nil {
		// best effort
		return
	} else if resp.StatusCode != 200 {
		// best effort
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	outputLines := strings.Split(string(body), "\n")

	lastRealLine := ""
	lastRealIndex := 0
	for i := len(job.lastTail) - 1; i >= 0; i-- {
		if len(job.lastTail[i]) > 0 {
			lastRealIndex = Max(i-compareWindow, 0)
			break
		}
	}
	for _, line := range job.lastTail[lastRealIndex:] {
		if len(line) > 0 {
			lastRealLine = line
			break
		}
		lastRealIndex += 1
	}
	startIndex := 0
	linesMatch := 0
	if lastRealIndex < tailLength-1 {
		for index, line := range outputLines {
			if line == lastRealLine {
				startIndex = index
				for _, tailLine := range job.lastTail[lastRealIndex:] {
					if startIndex >= len(outputLines) {
						startIndex = -1
						break
					} else if outputLines[startIndex] != tailLine {
						break
					} else {
						linesMatch += 1
						startIndex += 1
					}
				}
				if linesMatch == (tailLength - lastRealIndex) {
					break
				} else {
					linesMatch = 0
				}
			}
		}
	}
	if startIndex >= 0 {
		for _, line := range outputLines[startIndex:] {
			fmt.Println(line)
		}
	}
	tailIndex := tailLength - 1
	for i := len(outputLines) - 1; i >= Max(len(outputLines)-tailLength, 0); i-- {
		job.lastTail[tailIndex] = outputLines[i]
		tailIndex -= 1
	}

	return
}

func (job JarviceJob) Terminate() {
	logger.Ologger.Info("Terminating job " + job.Number + " (best effort)")
	http.PostForm(job.apiHost+"/jarvice/terminate", job.values)
	return
}

func (job JarviceJob) GoogleShutdownScript() string {
	return "#!/bin/bash \ncurl \"" + job.apiHost + "/jarvice/terminate?username=" + job.values.Get("username") +
		"&apikey=" + job.values.Get("apikey") + "&number=" + job.values.Get("number") + "\""
}
