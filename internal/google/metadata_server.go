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
	"io"
	"net/http"
	"strings"

	"jarvice.io/dragen/config"
	"jarvice.io/dragen/internal/logger"
)

func googleMetadata(path string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1"+path, nil)
	if err != nil {
		return "", err
	}
	req.Header = http.Header{
		"Metadata-Flavor": {"Google"},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ret := string(body)

	return ret, nil
}

func CheckDragenLicense() bool {
	dragen := config.DragenLic
	query, err := googleMetadata("/instance/licenses/")
	if err != nil {
		return false
	}
	for _, str := range strings.Split(strings.TrimSuffix(string(query), "\n"), "\n") {
		license, _ := googleMetadata("/instance/licenses/" + str + "id")
		if dragen == strings.TrimSuffix(license, "\n") {
			return true
		}
	}
	logger.Ologger.Warn("Unable to verify DRAGEN license")
	return false
}
