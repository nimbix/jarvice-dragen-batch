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

package google

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func googleMetadataServer() {
	go func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/computeMetadata/v1/project/project-id" {
				fmt.Fprint(w, "google-project")
			} else if path == "/computeMetadata/v1/instance/zone" {
				fmt.Fprint(w, "https://www.googleapis.com/compute/v1/projects/google-project/zones/us-central1-a")
			} else if path == "/computeMetadata/v1/instance/network-interfaces/0/network" {
				fmt.Fprint(w, "https://www.googleapis.com/compute/v1/projects/google-project/global/networks/default")
			} else if path == "/computeMetadata/v1/instance/name" {
				fmt.Fprint(w, "golang-test")
			} else if path == "/computeMetadata/v1/instance/id" {
				fmt.Fprint(w, "1234567890")
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer ts.Close()
	}()
}

func TestGoogleCompute(t *testing.T) {

	googleMetadataServer()
	time.Sleep(3 * time.Second)
	if vm, err := NewGoogleCompute(); err != nil {
		t.Error(err.Error())
	} else {
		if vm.GetName() != "golang-test" {
			t.Logf("%s\n", vm.GetName())
			t.Error("vm.GetName() failed")
		}
		if vm.GetProject() != "google-project" {
			t.Error("vm.GetProject() failed")
		}
		if vm.GetZone() != "us-central1-a" {
			t.Error("vm.GetZone() failed")
		}
		if vm.GetId() != "1234567890" {
			t.Error("vm.GetId() failed")
		}
	}
}
