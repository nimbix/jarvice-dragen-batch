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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

const (
	apiHost  = "http://localhost:8080"
	username = "jarvice"
	apikey   = "abc123"
	number   = "555"
	fnumber  = "777"
	nfnumber = "999"
)

var (
	job *JarviceJob = &JarviceJob{}
)

func checkApiArgs(values url.Values) bool {
	if values.Get("number") == "" {
		return false
	}
	if values.Get("username") == "" {
		return false
	}
	if values.Get("apikey") == "" {
		return false
	}
	return true
}

func checkApiArgsMachines(values url.Values) bool {
	if values.Get("username") == "" {
		return false
	}
	if values.Get("apikey") == "" {
		return false
	}
	return true
}

func jarviceServer(t *testing.T, terminate bool) *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if terminate && path != "/jarvice/terminate" {
			t.Error("Terminate() failed")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if path == "/jarvice/machines" {
			defer r.Body.Close()
			if body, err := io.ReadAll(r.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if query, perr := url.ParseQuery(string(body)); perr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					if !checkApiArgsMachines(query) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if query.Get("username") != username || query.Get("apikey") != apikey {
						w.WriteHeader(http.StatusUnauthorized)
						return
					} else {
						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		} else if path == "/jarvice/status" {
			defer r.Body.Close()
			if body, err := io.ReadAll(r.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if query, perr := url.ParseQuery(string(body)); perr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					if !checkApiArgs(query) {
						w.WriteHeader(http.StatusBadRequest)
						return
					} else {
						if query.Get("number") == number {
							jobStatusList := JobStatusList{
								number: JobStatus{
									Status: "PROCESSING STARTING",
								},
							}
							if resp, jerr := json.Marshal(jobStatusList); jerr != nil {
								w.WriteHeader(http.StatusInternalServerError)
								return
							} else {
								w.Write(resp)
								return
							}
						} else if query.Get("number") == fnumber {
							jobStatusList := JobStatusList{
								fnumber: JobStatus{
									Status: "COMPLETED",
								},
							}
							if resp, jerr := json.Marshal(jobStatusList); jerr != nil {
								w.WriteHeader(http.StatusInternalServerError)
								return
							} else {
								w.Write(resp)
								return
							}
						} else {
							w.WriteHeader(http.StatusNotFound)
							return
						}
					}
				}
			}
		} else if path == "/jarvice/terminate" {
			t.Log("terminate called successfully")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	go func(ts *httptest.Server) {
		ts.Start()
		time.Sleep(time.Hour)
	}(ts)
	return ts
}

func TestCheckAuth(t *testing.T) {
	job = NewJarviceJob(apiHost, username, apikey, number)
	ts := jarviceServer(t, false)
	if !job.CheckAuth() {
		t.Error("CheckAuth() failed")
	}
	ts.Close()
}

func TestCheckAuthUnauthorized(t *testing.T) {
	job = NewJarviceJob(apiHost, username, apikey+"123", nfnumber)
	ts := jarviceServer(t, false)
	if job.CheckAuth() {
		t.Error("CheckAuthUnauthorized() failed")
	}
	ts.Close()
}

func TestRunning(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, number)
	ts := jarviceServer(t, false)
	if !job.Running() {
		t.Error("Running() failed")
	}
	ts.Close()
}

func TestRunningNotFound(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, nfnumber)
	ts := jarviceServer(t, false)
	if job.Running() {
		t.Error("RunningNotFound() failed")
	}
	ts.Close()
}

func TestExitSuccess(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, fnumber)
	ts := jarviceServer(t, false)
	if !job.ExitSuccess() {
		t.Error("ExitSuccess() failed")
	}
	ts.Close()
}

func TestExitSuccessNotFound(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, nfnumber)
	ts := jarviceServer(t, false)
	if job.ExitSuccess() {
		t.Error("ExitSuccessNotFound() failed")
	}
	ts.Close()
}

func TestTerminate(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, number)
	ts := jarviceServer(t, true)
	job.Terminate()
	ts.Close()
}

func TestGoogleShutdownScript(t *testing.T) {
	job := NewJarviceJob(apiHost, username, apikey, number)
	if job.GoogleShutdownScript() != "#!/bin/bash \ncurl \""+apiHost+"/jarvice/terminate?username="+
		username+"&apikey="+apikey+"&number="+number+"\"" {
		t.Error("GoogleShutdownScript() failed")
	}
}
