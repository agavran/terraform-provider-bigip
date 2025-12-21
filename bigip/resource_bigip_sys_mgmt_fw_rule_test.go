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

var TEST_MGMT_FW_RULE_NAME = "test-rule"
var TEST_MGMT_FW_RULE_RESOURCE_NAME = "bigip_sys_mgmt_fw_rule." + TEST_MGMT_FW_RULE_NAME
var TEST_MGMT_FW_RULE_RESOURCE = `
resource "bigip_sys_mgmt_fw_rule" "test-rule" {
  name        = "` + TEST_MGMT_FW_RULE_NAME + `"
  action      = "accept"
	uuid        = "auto-generate"
  log         = "no"
  status      = "enabled"
  ip_protocol = "tcp"
  place_after = "first"
	description = "Firewall rule for testing purposes"
  source {
    addresses  { name = "30.40.40.0/24" }
		addresses  { name = "40.30.20.0/24" }
		addresses  { name = "13.33.55.10-13.33.55.250" }

    ports { name = "55" }
		ports { name = "8080" }
  }
  destination {
    addresses { name = "10.20.10.100" }
    ports { name = "9090" }
  }
}
`
var TEST_MGMT_FW_RULE_RESOURCE_UPDATE = `
resource "bigip_sys_mgmt_fw_rule" "test-rule" {
  name        = "` + TEST_MGMT_FW_RULE_NAME + `"
  action      = "accept"
	uuid        = "auto-generate"
  log         = "no"
  status      = "disabled"
  ip_protocol = "tcp"
  place_after = "first"
	description = "Updated rule for testing purposes"
  source {
    addresses  { name = "30.40.40.0/24" }
		addresses  { name = "40.30.20.0/24" }
    ports { name = "55" }
  }
  destination {
    addresses { name = "10.20.10.100" }
    ports { name = "9090" }
  }
}
`

func Test_Run_Multiple_Tests(t *testing.T) {
	t.Run("Create Test", TestAccBigipSysMgmtFwRule_Create)
	t.Run("Update Test", TestAccBigipSysMgmtFwRule_Update)
}

func TestAccBigipSysMgmtFwRule_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtFwRule_Destroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_FW_RULE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckMgmtFwRule_Exists(TEST_MGMT_FW_RULE_NAME, true),
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "name", TEST_MGMT_FW_RULE_NAME),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "action", "accept"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "status", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "ip_protocol", "tcp"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "description", "Firewall rule for testing purposes"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "place_after", "first"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "log", "no"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "source.0.addresses.1.name", "30.40.40.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "source.0.addresses.2.name", "40.30.20.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "source.0.addresses.0.name", "13.33.55.10-13.33.55.250"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "source.0.ports.0.name", "55"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "source.0.ports.1.name", "8080"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "destination.0.addresses.0.name", "10.20.10.100"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "destination.0.ports.0.name", "9090"),
				),
			},
		},
	})
}
func TestAccBigipSysMgmtFwRule_Update(t *testing.T) {
	var firstUUID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtFwRule_Destroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_FW_RULE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					//testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "status", "enabled"),
					testCaptureUUID(TEST_MGMT_FW_RULE_NAME, &firstUUID),
				),
			},
			{
				Config: TEST_MGMT_FW_RULE_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					//testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_fw_rule.test-rule", "status", "disabled"),
					testCheckUUIDPreserved(TEST_MGMT_FW_RULE_NAME, &firstUUID),
				),
			},
		},
	})
}
func TestAccBigipSysMgmtFwRule_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtFwRule_Destroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_FW_RULE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					testCheckMgmtFwRule_Exists(TEST_MGMT_FW_RULE_NAME, true),
				),
			},
			{
				ResourceName:      TEST_MGMT_FW_RULE_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"place_after",  // Placement is not stored, only reflected in rule order
					"place_before", // Placement is not stored, only reflected in rule order
					"uuid",         // UUID may be "auto-generate" in config but actual UUID in state
				},
				ImportStateId:    TEST_MGMT_FW_RULE_NAME,
				ImportStateCheck: testPrintImportedState,
			},
		},
	})
}

func testCheckMgmtFwRule_Exists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		rule, err := client.GetManagementFwRule(name)
		if err != nil {
			return err
		}
		if exists && rule == nil {
			return fmt.Errorf("rule %s was not created.", name)
		}
		if !exists && rule != nil {
			return fmt.Errorf("rule %s still exists.", name)
		}
		return nil
	}
}

func testCheckMgmtFwRule_Destroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_mgmt_fw_rule" {
			continue
		}

		name := rs.Primary.ID
		rule, err := client.GetManagementFwRule(name)
		if err != nil {
			return err
		}
		if rule != nil {
			return fmt.Errorf("rule %s not destroyed.", name)
		}
		fmt.Printf("[INFO] Management firewall rule '%s' has been successfully destroyed\n", name)
	}
	return nil
}

func testCaptureUUID(name string, capturedUUID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		rule, err := client.GetManagementFwRule(name)
		if err != nil {
			return err
		}
		if rule == nil {
			return fmt.Errorf("rule %s not found", name)
		}

		*capturedUUID = rule.UUID
		fmt.Printf("[INFO] Captured UUID: %s\n", *capturedUUID)
		return nil
	}
}

func testCheckUUIDPreserved(name string, expectedUUID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		rule, err := client.GetManagementFwRule(name)
		if err != nil {
			return err
		}
		if rule == nil {
			return fmt.Errorf("rule %s not found", name)
		}

		if rule.UUID != *expectedUUID {
			return fmt.Errorf("UUID changed during update: expected %s, got %s", *expectedUUID, rule.UUID)
		}

		fmt.Printf("[INFO] UUID preserved: %s\n", rule.UUID)
		return nil
	}
}
