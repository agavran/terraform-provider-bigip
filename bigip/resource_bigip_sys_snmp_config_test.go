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

var TEST_SNMP_CONFIG_NAME = "snmp_config"
var TEST_SNMP_CONFIG_RESOURCE_NAME = "bigip_sys_snmp_config.test-snmp-config"
var TEST_SNMP_CONFIG_RESOURCE = `
resource "bigip_sys_snmp_config" "test-snmp-config" {
	sys_contact       = "Test Admin <test@example.com>"
	sys_location      = "Test Server Room"
	allowed_addresses = ["10.0.0.0/8"]
}
`
var TEST_SNMP_CONFIG_RESOURCE_UPDATE = `
resource "bigip_sys_snmp_config" "test-snmp-config" {
	sys_contact       = "Updated Admin <admin@example.com>"
	sys_location      = "Production Data Center"
	allowed_addresses = ["10.0.0.0/8", "192.168.1.0/24"]
}
`

func TestAccBigipSysSnmpConfig_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpConfigExists(true),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "sys_contact", "Test Admin <test@example.com>"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "sys_location", "Test Server Room"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "allowed_addresses.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "allowed_addresses.0", "10.0.0.0/8"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpConfig_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpConfigDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "sys_contact", "Test Admin <test@example.com>"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "allowed_addresses.#", "1"),
				),
			},
			{
				Config: TEST_SNMP_CONFIG_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "sys_contact", "Updated Admin <admin@example.com>"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "sys_location", "Production Data Center"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "allowed_addresses.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_config.test-snmp-config", "allowed_addresses.1", "192.168.1.0/24"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpConfig_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpConfigDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpConfigExists(true),
				),
			},
			{
				ResourceName:      TEST_SNMP_CONFIG_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SNMP_CONFIG_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckSnmpConfigExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetSnmpConfig()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("SNMP configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("SNMP configuration still exists.")
		}
		return nil
	}
}

func testCheckSnmpConfigDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_snmp_config" {
			continue
		}

		snmpConfig, err := client.GetSnmpConfig()
		if err != nil {
			return err
		}

		var errors []string

		if snmpConfig.SysContact != "Customer Name <admin@customer.com>" {
			errors = append(errors, fmt.Sprintf("sys_contact not reset to default (Customer Name <admin@customer.com>), got: %s", snmpConfig.SysContact))
		}
		if snmpConfig.SysLocation != "Network Closet 1" {
			errors = append(errors, fmt.Sprintf("sys_location not reset to default (Network Closet 1), got: %s", snmpConfig.SysLocation))
		}
		if len(snmpConfig.AllowedAddresses) != 1 || snmpConfig.AllowedAddresses[0] != "127.0.0.0/8" {
			errors = append(errors, fmt.Sprintf("allowed_addresses not reset to default ([127.0.0.0/8]), got: %v", snmpConfig.AllowedAddresses))
		}

		if len(errors) > 0 {
			errorMsg := "SNMP configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] SNMP configuration successfully reset to default values")
	}
	return nil
}
