/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestPartition = "Common"

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.

// var testAccProviders map[string]*schema.Provider{}
// var testAccProvider *schema.Provider

var testAccProvider = Provider()
var testAccProviders = map[string]*schema.Provider{
	"bigip": testAccProvider,
}

func TestAccProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAcctPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if os.Getenv("BIGIP_TOKEN_VALUE") != "" || (os.Getenv("BIGIP_TOKEN_AUTH") != "" && os.Getenv("BIGIP_LOGIN_REF") != "") {
		return
	}
	if v := os.Getenv("BIGIP_TEST_PARTITION"); v != "" {
		TestPartition = v
	}
	for _, s := range [...]string{"BIGIP_HOST", "BIGIP_USER", "BIGIP_PASSWORD"} {
		if os.Getenv(s) == "" {
			t.Fatal("Either BIGIP_TOKEN_AUTH + BIGIP_LOGIN_REF or BIGIP_USER, BIGIP_PASSWORD and BIGIP_HOST are required for tests.")
			return
		}
	}
}

func testAcctUnitPreCheck(_ *testing.T, url string) {
	_ = os.Setenv("BIGIP_HOST", url)
	_ = os.Setenv("BIGIP_USER", "xxxx")
	_ = os.Setenv("BIGIP_PASSWORD", "xxx")
	_ = os.Setenv("BIGIP_TOKEN_AUTH", "false")
}

// loadFixtureBytes returns the entire contents of the given file as a byte slice
func loadFixtureBytes(path string) []byte {
	contents, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return contents
}

// loadFixtureString returns the entire contents of the given file as a string
func loadFixtureString(path string) string {
	return string(loadFixtureBytes(path))
}

// testCheckSleep is a helper that pauses test execution for a specified duration.
func testCheckSleep(seconds ...int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		timer := 60
		if len(seconds) > 0 && seconds[0] > 0 {
			timer = seconds[0]
		}
		fmt.Printf("Pausing for %d seconds...\n", timer)
		time.Sleep(time.Duration(timer) * time.Second)
		return nil
	}
}

func testPrintImportedState(s []*terraform.InstanceState) error {
	if len(s) != 1 {
		return fmt.Errorf("expected 1 state, got %d", len(s))
	}
	state := s[0]
	fmt.Printf("\n=== IMPORTED STATE ===\n")
	fmt.Printf("ID: %s\n", state.ID)
	fmt.Printf("Attributes:\n")
	for k, v := range state.Attributes {
		fmt.Printf("  %s = %s\n", k, v)
	}
	fmt.Printf("======================\n\n")
	return nil
}
