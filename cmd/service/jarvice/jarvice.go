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

package jarvice

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"jarvice.io/dragen/config"
)

type DragenParams struct {
	Command      string `json:"command"`
	GcpVmid      string `json:"GCP_VMID"`
	GcpProjectid string `json:"GCP_PROJECTID"`
	GcpZone      string `json:"GCP_ZONE"`
}

type DragenApp struct {
	Command    string       `json:"command"`
	Geometry   string       `json:"geometry"`
	Parameters DragenParams `json:"parameters"`
}

type Machine struct {
	Type  string `json:"type"`
	Nodes int    `json:"nodes"`
}

type Vault struct {
	Name     string `json:"name"`
	Readonly bool   `json:"readonly"`
	Force    bool   `json:"force"`
}

type User struct {
	Username string `json:"username"`
	Apikey   string `json:"apikey"`
}

type JobSubmission struct {
	App         string    `json:"app"`
	Staging     bool      `json:"staging"`
	JobLabel    string    `json:"job_label"`
	Application DragenApp `json:"application"`
	Machine     Machine   `json:"machine"`
	Vault       Vault     `json:"vault"`
	User        User      `json:"user"`
	Priority    string    `json:"job_priority"`
}

type JobResponse struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
}

func SubmitJarviceJob(app, machine, vmid, project, zone,
	apiHost, username, apikey, priority, base64Args string) (string, error) {

	values := &JobSubmission{
		App:     app,
		Staging: false,
		Application: DragenApp{
			Command:  "Batch",
			Geometry: "1,920x1,080",
			Parameters: DragenParams{
				Command: "MYDRAGEN_ARGS=$(echo " +
					base64Args +
					" | base64 -d); " + config.DragenCommand + " $MYDRAGEN_ARGS",
				GcpVmid:      vmid,
				GcpProjectid: project,
				GcpZone:      zone,
			},
		},
		Machine: Machine{
			Type:  machine,
			Nodes: 1,
		},
		Vault: Vault{
			Name:     "ephemeral",
			Readonly: false,
			Force:    false,
		},
		User: User{
			Username: username,
			Apikey:   apikey,
		},
		Priority: priority,
	}
	values.JobLabel = "vmid=" + vmid

	jsonBlob, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(apiHost+"/jarvice/submit", "application/json", bytes.NewBuffer(jsonBlob))
	if err != nil {
		return "", err
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", errors.New("JARVICE job submission failed")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	jobResponse := JobResponse{}
	json.Unmarshal(body, &jobResponse)
	return strconv.Itoa(jobResponse.Number), nil
}
