/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package egress

import (
	_ "fmt"
	"github.com/spf13/cobra"
	_ "io"
	"net/http"
	"time"

	"github.com/openshift/rosa/pkg/arguments"
	"github.com/openshift/rosa/pkg/rosa"
)

var Cmd = &cobra.Command{
	Use:   "egress",
	Short: "Verify AWS egress is ok for cluster install",
	Long:  "Verify AWS egress checks whether DNS and egress access works to a set of domains/urls.",
	Example: `  # Verify AWS egress is configured correctly
  rosa verify egress`,
	RunE: run,
}
var verbose bool = false
var quiet bool = false

func init() {
	flags := Cmd.Flags()
	// add a flag for verbose output
	flags.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	// add a flag for quiet mode
	flags.BoolVar(&quiet, "quiet", false, "Enable quiet mode")

	arguments.AddProfileFlag(flags)
}

func run(cmd *cobra.Command, _ []string) (err error) {
	r := rosa.NewRuntime().WithOCM()
	defer r.Cleanup()

	var client = &http.Client{
		//CheckRedirect: redirectPolicyFunc,
		Timeout: 5 * time.Second,
	}

	// Create an array of urls to test
	urls := []string{
		"https://registry.redhat.io",
		"https://quay.io",
		"https://sso.redhat.com",
		"https://console.redhat.com/openshift",
		"https://quay-registry.s3.amazonaws.com",
		"https://cm-quay-production-s3.s3.amazonaws.com",
		"https://cart-rhcos-ci.s3.amazonaws.com",
		"https://openshift.org",
		"https://registry.access.redhat.com",
		"https://console.redhat.com",
		"https://pull.q1w2.quay.rhcloud.com",
		"https://q1w2.quay.rhcloud.com",
		// telemetry
		"https://cert-api.access.redhat.com",
		"https://api.access.redhat.com",
		"https://infogw.api.openshift.com",
		"https://console.redhat.com",
		"https://observatorium.api.openshift.com",
		// mirrors
		"https://mirror.openshift.com",
		"https://storage.googleapis.com/openshift-release",
		"https://api.openshift.com",
		// sre
		"https://api.pagerduty.com/",
		"https://events.pagerduty.com",
		"https://api.deadmanssnitch.com",
		"https://nosnch.in",
		"https://splunkcloud.com", // hack, there were a dozen variations of this url
	}

	type result struct {
		url    string
		status int
	}
	// create an array of the same size as the urls array to hold the results
	responses := make([]result, len(urls))

	// Loop over urls
	for i, url := range urls {
		responses[i].url = url
		resp, err := client.Get(url)
		if err != nil {
			r.Reporter.Errorf("Error getting URL %v\n%v", url, err)
			responses[i].status = 400
		} else {
			responses[i].status = resp.StatusCode
			//defer resp.Body.Close()
			//_, err = io.ReadAll(resp.Body)
			//if err != nil {
			//	r.Reporter.Errorf("Error reading response body: %v", err)
			//}
		}
	}

	// Loop over responses
	r.Reporter.Infof("*** Summary Report on egress access for %v URLs ***", len(urls))
	for _, response := range responses {
		if response.status != 200 {
			r.Reporter.Errorf("Error: GET: %v, status: %v", response.url, response.status)
		} else {
			if verbose {
				r.Reporter.Infof("OK: GET: %v, status: %v", response.url, response.status)
			}
		}
	}
	return nil
}
