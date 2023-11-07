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
	"os"

	"github.com/spf13/cobra"
	"jarvice.io/dragen/cmd/service/batch"
	"jarvice.io/dragen/config"
	"jarvice.io/dragen/internal/logger"
	"jarvice.io/dragen/internal/monitor"
)

var (
	bflag          bool
	apiHost        string
	username       string
	apikey         string
	machine        string
	s3AccessKey    string
	s3SecretKey    string
	dragenApp      string
	illuminaLic    string
	serviceAccount string
	priority       string

	rootCmd = &cobra.Command{
		Use:   "service",
		Short: "Batch service for Dragen JARVICE job.",
		Long:  `Batch service for Dragen JARVICE job.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("build").Changed {
				logger.Ologger.Info("build: " + config.Build)
				os.Exit(0)
			}
			return
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dragenBatch, err := batch.NewDragenBatch(apiHost, username, apikey,
				dragenApp, machine, s3AccessKey, s3SecretKey, illuminaLic,
				serviceAccount, priority, args...)
			if err != nil {
				return err
			} else {
				if err := monitor.StartMonitor(dragenBatch); err != nil {
					return err
				}
			}
			return nil
		},
		Version:       config.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return rootCmd.Execute()
}
func init() {
	illuminaLic = os.Getenv("ILLUMINA_LIC_SERVER")
	rootCmd.Flags().StringVar(&apiHost, "api-host", config.JarviceApi, "JARVICE API URL")
	rootCmd.Flags().StringVar(&machine, "machine", config.JarviceMachine, "JARVICE machine type")
	rootCmd.Flags().StringVar(&username, "username", os.Getenv("JARVICE_API_USER"), "JARVICE API username")
	rootCmd.Flags().StringVar(&apikey, "apikey", os.Getenv("JARVICE_API_KEY"), "JARVICE apikey")
	rootCmd.Flags().StringVar(&s3AccessKey, "s3-access-key", os.Getenv("S3_ACCESS_KEY"), "s3 access key")
	rootCmd.Flags().StringVar(&s3SecretKey, "s3-secret-key", os.Getenv("S3_SECRET_KEY"), "s3 secret key")
	rootCmd.Flags().StringVar(&dragenApp, "dragen-app", "", "Dragen JARVICE application")
	rootCmd.Flags().BoolVar(&bflag, "build", false, "Build info")
	rootCmd.Flags().StringVar(&serviceAccount, "google-sa", "default", "Google Cloud service account")
	rootCmd.Flags().StringVar(&priority, "job-priority", "normal", "JARVICE job priority")
	rootCmd.MarkFlagRequired("dragen-app")
}
