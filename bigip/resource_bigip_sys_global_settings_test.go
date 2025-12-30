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

var TEST_GLOBAL_SETTINGS_NAME = "global_settings"
var TEST_GLOBAL_SETTINGS_RESOURCE_NAME = "bigip_sys_global_settings.test-global-settings"
var TEST_GLOBAL_SETTINGS_RESOURCE = `
resource "bigip_sys_global_settings" "test-global-settings" {
	gui_security_banner      = "enabled"
	gui_security_banner_text = "Test Banner - Authorized Access Only"
	hostname                 = "test-bigip.lab.local"
}
`
var TEST_GLOBAL_SETTINGS_RESOURCE_UPDATE = `
resource "bigip_sys_global_settings" "test-global-settings" {
	gui_security_banner      = "disabled"
	gui_security_banner_text = "Updated Banner Text"
	hostname                 = "prod-bigip-01.lab.local"
}
`

func TestAccBigipSysGlobalSettings_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_GLOBAL_SETTINGS_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckGlobalSettingsExists(true),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner_text", "Test Banner - Authorized Access Only"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "hostname", "test-bigip.lab.local"),
				),
			},
		},
	})
}

func TestAccBigipSysGlobalSettings_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckGlobalSettingsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_GLOBAL_SETTINGS_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "hostname", "test-bigip.lab.local"),
				),
			},
			{
				Config: TEST_GLOBAL_SETTINGS_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner_text", "Updated Banner Text"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "hostname", "prod-bigip-01.lab.local"),
				),
			},
		},
	})
}

func TestAccBigipSysGlobalSettings_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckGlobalSettingsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_GLOBAL_SETTINGS_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckGlobalSettingsExists(true),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "gui_security_banner_text", "Test Banner - Authorized Access Only"),
					resource.TestCheckResourceAttr("bigip_sys_global_settings.test-global-settings", "hostname", "test-bigip.lab.local"),
				),
			},
			{
				ResourceName:      TEST_GLOBAL_SETTINGS_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_GLOBAL_SETTINGS_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckGlobalSettingsExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetGlobalSettings()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("Global Settings configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("Global Settings configuration still exists.")
		}
		return nil
	}
}

func testCheckGlobalSettingsDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_global_settings" {
			continue
		}

		globalSettings, err := client.GetGlobalSettings()
		if err != nil {
			return err
		}

		var errors []string

		if globalSettings.Hostname != "bigip1" {
			errors = append(errors, fmt.Sprintf("hostname not reset to default (bigip1), got: %s", globalSettings.Hostname))
		}
		if globalSettings.GuiSecurityBanner != "enabled" {
			errors = append(errors, fmt.Sprintf("gui_security_banner not reset to default (enabled), got: %s", globalSettings.GuiSecurityBanner))
		}
		if globalSettings.GuiSecurityBannerText != "Welcome to the BIG-IP Configuration Utility.\n\nLog in with your username and password using the fields on the left." {
			errors = append(errors, fmt.Sprintf("gui_security_banner_text not reset to default, got: %s", globalSettings.GuiSecurityBannerText))
		}

		if len(errors) > 0 {
			errorMsg := "Global Settings configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] Global Settings configuration successfully reset to default values")
	}
	return nil
}
