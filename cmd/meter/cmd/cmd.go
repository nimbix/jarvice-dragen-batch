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

package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"jarvice.io/dragen/cmd/meter/dragen"
	"jarvice.io/dragen/config"
	"jarvice.io/dragen/internal/google"
	"jarvice.io/dragen/internal/monitor"
)

var (
	bflag    bool
	apiHost  string
	username string
	apikey   string
	jobId    string
	service  string
	rootCmd  = &cobra.Command{
		Use:   "meter",
		Short: "A meter service for Dragen JARVICE job.",
		Long:  `A meter service for Dragen JARVICE job.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("build").Changed {
				slog.Info("build: " + config.Build)
				os.Exit(0)
			}
			return
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if meter, err := dragen.NewDragenMeter(apiHost, username, apikey, jobId, service); err != nil {
				return err
			} else {
				if err := monitor.StartMonitor(meter); err != nil {
					return err
				}
			}
			return nil
		},
		Version: config.Version,
	}
)

func DeleteHost() {
	if vm, err := google.NewGoogleCompute(); err == nil {
		slog.Warn("Google virtual machine hosting meter service being removed")
		vm.DeleteHost()
	}
	return
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
func init() {
	rootCmd.Flags().StringVar(&apiHost, "api-host", config.JarviceApi, "JARVICE API URL")
	rootCmd.Flags().StringVar(&username, "username", "", "JARVICE API username")
	rootCmd.Flags().StringVar(&apikey, "apikey", "", "JARVICE apikey")
	rootCmd.Flags().StringVar(&jobId, "job-id", "", "JARVICE job ID")
	rootCmd.Flags().BoolVar(&bflag, "build", false, "Build info")
	rootCmd.Flags().StringVar(&service, "service-name", "", "Google Batch service")
	rootCmd.MarkFlagRequired("username")
	rootCmd.MarkFlagRequired("apikey")
	rootCmd.MarkFlagRequired("job-id")
	rootCmd.MarkFlagRequired("service-name")
}
