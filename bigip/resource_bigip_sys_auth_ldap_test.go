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

var TEST_LDAP_CONFIG_NAME = "system-auth"
var TEST_LDAP_CONFIG_RESOURCE_NAME = "bigip_sys_auth_ldap.test-ldap"
var TEST_LDAP_CONFIG_RESOURCE = `

resource "bigip_sys_auth_ldap" "test-ldap" {
	name              = "` + TEST_LDAP_CONFIG_NAME + `"
	servers           = ["192.168.120.62"]
	port              = 389
	bind_dn           = "cn=ldap_user,dc=test,dc=com"
	bind_pw           = "password123"
	search_base_dn    = "dc=test,dc=com"
	login_attribute   = "uid"
	ssl               = "disabled"
	scope             = "sub"
	referrals         = "no"
	check_roles_group = "enabled"
}
`

var TEST_LDAP_CONFIG_RESOURCE_UPDATE = `

resource "bigip_sys_auth_ldap" "test-ldap" {
	name              = "` + TEST_LDAP_CONFIG_NAME + `"
	servers           = ["192.168.120.62", "192.168.120.63"]
	port              = 389
	bind_dn           = "cn=ldap_user,dc=test,dc=com"
	bind_pw           = "password123"
	search_base_dn    = "dc=test,dc=com"
	login_attribute   = "uid"
	ssl               = "disabled"
	scope             = "sub"
	referrals         = "no"
	check_roles_group = "enabled"
	debug             = "enabled"
	warnings = "disabled"
}
`

func TestAccBigipSysAuthLdap_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckLdapConfigDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_LDAP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckLdapConfigExists(TEST_LDAP_CONFIG_NAME, true),
					//testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "name", "system-auth"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.0", "192.168.120.62"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "port", "389"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "bind_dn", "cn=ldap_user,dc=test,dc=com"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "search_base_dn", "dc=test,dc=com"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "login_attribute", "uid"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "ssl", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "scope", "sub"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "referrals", "no"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "check_roles_group", "enabled"),
				),
			},
		},
	})
}

func TestAccBigipSysAuthLdap_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckLdapConfigDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_LDAP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "port", "389"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "ssl", "disabled"),
				),
			},
			{
				Config: TEST_LDAP_CONFIG_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.0", "192.168.120.62"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "servers.1", "192.168.120.63"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "port", "389"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "login_attribute", "uid"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "check_roles_group", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "ssl", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "scope", "sub"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "referrals", "no"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "debug", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_auth_ldap.test-ldap", "warnings", "disabled"),
				),
			},
		},
	})
}

func TestAccBigipSysAuthLdap_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckLdapConfigDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_LDAP_CONFIG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckLdapConfigExists(TEST_LDAP_CONFIG_NAME, true),
				),
			},
			{
				ResourceName:            TEST_LDAP_CONFIG_RESOURCE_NAME,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bind_pw"}, // Password is returned encrypted
				ImportStateId:           TEST_LDAP_CONFIG_NAME,
				ImportStateCheck:        testPrintImportedState,
			},
		},
	})
}

func testCheckLdapConfigExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetLdapConfig(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("LDAP config %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("LDAP config %s still exists.", name)
		}
		return nil
	}
}

func testCheckLdapConfigDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_auth_ldap" {
			continue
		}

		name := rs.Primary.ID
		ldapConfig, err := client.GetLdapConfig(name)
		if err != nil {
			return err
		}
		if ldapConfig != nil {
			return fmt.Errorf("LDAP config %s not destroyed.", name)
		}
		fmt.Printf("[INFO] System Auth LDAP configuration '%s' has been successfully destroyed\n", name)
	}
	return nil
}
