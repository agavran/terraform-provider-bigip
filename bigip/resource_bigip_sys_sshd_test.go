/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"fmt"
	"testing"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TEST_SSHD__CONFIG_NAME = "sshd_config"
var TEST_SSHD_RESOURCE_NAME = "bigip_sys_sshd.test-sshd"
var TEST_SSHD_RESOURCE = `
resource "bigip_sys_sshd" "test-sshd" {
	allow              = ["10.0.0.0/8", "192.168.1.0/24"]
	banner             = "enabled"
	banner_text        = "Test SSH Banner"
	inactivity_timeout = 300
	log_level          = "verbose"
	port               = 22
}
`
var TEST_SSHD_RESOURCE_UPDATE = `
resource "bigip_sys_sshd" "test-sshd" {
	allow              = ["172.16.0.0/12"]
	banner             = "disabled"
	banner_text        = "none"
	inactivity_timeout = 600
	log_level          = "debug"
	port               = 2222
}
`

func TestAccBigipSysSshd_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SSHD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSshdExists(true),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "allow.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "allow.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "allow.1", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "banner", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "banner_text", "Test SSH Banner"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "inactivity_timeout", "300"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "log_level", "verbose"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "port", "22"),
				),
			},
		},
	})
}

func TestAccBigipSysSshd_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSshdDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SSHD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "banner", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "port", "22"),
				),
			},
			{
				Config: TEST_SSHD_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "allow.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "allow.0", "172.16.0.0/12"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "banner", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "inactivity_timeout", "600"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "log_level", "debug"),
					resource.TestCheckResourceAttr("bigip_sys_sshd.test-sshd", "port", "2222"),
				),
			},
		},
	})
}

func TestAccBigipSysSshd_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSshdDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SSHD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSshdExists(true),
				),
			},
			{
				ResourceName:      TEST_SSHD_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SSHD__CONFIG_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckSshdExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetSSHDConfig()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("SSHD configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("SSHD configuration still exists.")
		}
		return nil
	}
}

func testCheckSshdDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_sshd" {
			continue
		}

		sshdConfig, err := client.GetSSHDConfig()
		if err != nil {
			return err
		}

		var errors []string

		if sshdConfig.Banner != "disabled" {
			errors = append(errors, fmt.Sprintf("banner not reset to default (disabled), got: %s", sshdConfig.Banner))
		}

		if sshdConfig.BannerText != "" && sshdConfig.BannerText != "none" {
			errors = append(errors, fmt.Sprintf("banner_text present with non-default value (expected none or omitted), got: %s", sshdConfig.BannerText))
		}
		if sshdConfig.InactivityTimeout != 0 {
			errors = append(errors, fmt.Sprintf("inactivity_timeout not reset to default (0), got: %d", sshdConfig.InactivityTimeout))
		}
		if sshdConfig.Include != "" && sshdConfig.Include != "none" {
			errors = append(errors, fmt.Sprintf("include present with non-default value (expected none or omitted), got: %s", sshdConfig.Include))
		}
		if sshdConfig.LogLevel != "info" {
			errors = append(errors, fmt.Sprintf("log_level not reset to default (info), got: %s", sshdConfig.LogLevel))
		}
		if sshdConfig.Login != "enabled" {
			errors = append(errors, fmt.Sprintf("login not reset to default (enabled), got: %s", sshdConfig.Login))
		}
		if sshdConfig.Port != 22 {
			errors = append(errors, fmt.Sprintf("port not reset to default (22), got: %d", sshdConfig.Port))
		}

		if len(errors) > 0 {
			errorMsg := "SSHD configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] SSHD configuration successfully reset to default values")
	}
	return nil
}
