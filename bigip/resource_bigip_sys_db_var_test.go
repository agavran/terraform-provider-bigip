/*
Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"fmt"
	"testing"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TEST_SYS_DB_VAR_NAME = "ui.advisory.enabled"
var TEST_SYS_DB_VAR_RESOURCE_NAME = "bigip_sys_db_variable.test-db-var"
var TEST_SYS_DB_VAR_RESOURCE = `

resource "bigip_sys_db_variable" "test-db-var" {
  name        = "ui.advisory.enabled"
  value      = "true"
}
`
var TEST_SYS_DB_VAR_RESOURCE_UPDATE = `

resource "bigip_sys_db_variable" "test-db-var" {
  name        = "ui.advisory.enabled"
  value      = "false"
}
`

func Test_Run_Multiple_Tests_DB_Var(t *testing.T) {
	t.Run("Create Test", TestAccBigipSysDbVar_Create)
	t.Run("Update Test", TestAccBigipSysDbVar_Update)
}

func TestAccBigipSysDbVar_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSysDbVar_ResetToDefaults,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYS_DB_VAR_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_db_variable.test-db-var", "name", "ui.advisory.enabled"),
					resource.TestCheckResourceAttr("bigip_sys_db_variable.test-db-var", "value", "true"),
				),
			},
		},
	})
}
func TestAccBigipSysDbVar_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYS_DB_VAR_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_db_variable.test-db-var", "value", "true"),
				),
			},
			{
				Config: TEST_SYS_DB_VAR_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_db_variable.test-db-var", "value", "false"),
				),
			},
		},
	})
}

func TestAccBigipSysDbVar_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSysDbVar_ResetToDefaults,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYS_DB_VAR_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSysDbVar_Exists(TEST_SYS_DB_VAR_NAME, true),
					resource.TestCheckResourceAttr("bigip_sys_db_variable.test-db-var", "value", "true"),
				),
			},
			{
				ResourceName:      TEST_SYS_DB_VAR_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SYS_DB_VAR_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckSysDbVar_Exists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		sysDb, err := client.GetDBVariable(name)
		if err != nil {
			return err
		}
		if exists && sysDb == nil {
			return fmt.Errorf("DB variable %s exists on the system", name)
		}
		if !exists && sysDb != nil {
			return fmt.Errorf("DB variable %s does not exist on the system", name)
		}
		return nil
	}
}

func testCheckSysDbVar_ResetToDefaults(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_db_variable" {
			continue
		}

		name := rs.Primary.ID
		sysDb, err := client.GetDBVariable(name)
		if err != nil {
			return err
		}
		defaultVal, ok := bigip.DefaultDBValues[name]
		if !ok {
			return fmt.Errorf("default value not found for DB variable '%s'. This variable may not be supported or may not have a default value defined", name)
		}

		if sysDb.Value == defaultVal {
			fmt.Printf("[INFO] Default value for DB variable '%s' has been set correctly to '%s'\n", name, defaultVal)
			return nil
		} else {
			return fmt.Errorf("value '%s' set for DB variable '%s' on the BIGIP system does not match defaul value '%s'", sysDb.Value, name, defaultVal)
		}
	}
	return nil
}
