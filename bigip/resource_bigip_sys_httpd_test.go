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

var TEST_HTTPD_NAME = "httpd_config"
var TEST_HTTPD_RESOURCE_NAME = "bigip_sys_httpd.test-httpd"
var TEST_HTTPD_RESOURCE = `
resource "bigip_sys_httpd" "test-httpd" {
	allow                   = ["10.0.0.0/8", "172.18.0.0/16"]
	auth_pam_idle_timeout   = 900
	max_clients             = 15
	log_level               = "info"
}
`

var TEST_HTTPD_RESOURCE_UPDATE = `
resource "bigip_sys_httpd" "test-httpd" {
	allow                   = ["192.168.1.0/24", "172.0.0.0/8"]
	auth_pam_idle_timeout   = 1800
	max_clients             = 20
	log_level               = "warn"
}
`

func TestAccBigipSysHttpd_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_HTTPD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckHttpdExists(true),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.1", "172.18.0.0/16"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "auth_pam_idle_timeout", "900"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "max_clients", "15"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "log_level", "info"),
				),
			},
		},
	})
}

func TestAccBigipSysHttpd_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckHttpdDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_HTTPD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "auth_pam_idle_timeout", "900"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "max_clients", "15"),
				),
			},
			{
				Config: TEST_HTTPD_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "allow.1", "172.0.0.0/8"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "auth_pam_idle_timeout", "1800"),
					resource.TestCheckResourceAttr("bigip_sys_httpd.test-httpd", "max_clients", "20"),
				),
			},
		},
	})
}

func TestAccBigipSysHttpd_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckHttpdDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_HTTPD_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckHttpdExists(true),
				),
			},
			{
				ResourceName:      TEST_HTTPD_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_HTTPD_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckHttpdExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetHTTPDConfig()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("HTTPD configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("HTTPD configuration still exists.")
		}
		return nil
	}
}

func testCheckHttpdDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_httpd" {
			continue
		}

		httpdConfig, err := client.GetHTTPDConfig()
		if err != nil {
			return err
		}

		var errors []string

		if len(httpdConfig.Allow) != 1 || httpdConfig.Allow[0] != "All" {
			errors = append(errors, fmt.Sprintf("allow not reset to default ([All]), got: %v", httpdConfig.Allow))
		}

		if httpdConfig.AuthName != "BIG-IP" {
			errors = append(errors, fmt.Sprintf("auth_name not reset to default (BIG-IP), got: %s", httpdConfig.AuthName))
		}
		if httpdConfig.AuthPamDashboardTimeout != "off" {
			errors = append(errors, fmt.Sprintf("auth_pam_dashboard_timeout not reset to default (off), got: %s", httpdConfig.AuthPamDashboardTimeout))
		}
		if httpdConfig.AuthPamIdleTimeout != 1200 {
			errors = append(errors, fmt.Sprintf("auth_pam_idle_timeout not reset to default (1200), got: %d", httpdConfig.AuthPamIdleTimeout))
		}
		if httpdConfig.AuthPamValidateIp != "on" {
			errors = append(errors, fmt.Sprintf("auth_pam_validate_ip not reset to default (on), got: %s", httpdConfig.AuthPamValidateIp))
		}
		if httpdConfig.FastcgiTimeout != 300 {
			errors = append(errors, fmt.Sprintf("fastcgi_timeout not reset to default (300), got: %d", httpdConfig.FastcgiTimeout))
		}
		if httpdConfig.HostnameLookup != "off" {
			errors = append(errors, fmt.Sprintf("hostname_lookup not reset to default (off), got: %s", httpdConfig.HostnameLookup))
		}
		// Include, ssl_ca_cert_file, ssl_certchainfile, ssl_include should be "none" or empty
		if httpdConfig.Include != "" && httpdConfig.Include != "none" {
			errors = append(errors, fmt.Sprintf("include present with non-default value (expected none or omitted), got: %s", httpdConfig.Include))
		}
		if httpdConfig.LogLevel != "warn" {
			errors = append(errors, fmt.Sprintf("log_level not reset to default (warn), got: %s", httpdConfig.LogLevel))
		}
		if httpdConfig.MaxClients != 10 {
			errors = append(errors, fmt.Sprintf("max_clients not reset to default (10), got: %d", httpdConfig.MaxClients))
		}
		if httpdConfig.RedirectHttpToHttps != "disabled" {
			errors = append(errors, fmt.Sprintf("redirect_http_to_https not reset to default (disabled), got: %s", httpdConfig.RedirectHttpToHttps))
		}
		if httpdConfig.SslPort != 443 {
			errors = append(errors, fmt.Sprintf("ssl_port not reset to default (443), got: %d", httpdConfig.SslPort))
		}
		if httpdConfig.SslCaCertFile != "" && httpdConfig.SslCaCertFile != "none" {
			errors = append(errors, fmt.Sprintf("ssl_ca_cert_file present with non-default value (expected none or omitted), got: %s", httpdConfig.SslCaCertFile))
		}
		if httpdConfig.SslCertchainfile != "" && httpdConfig.SslCertchainfile != "none" {
			errors = append(errors, fmt.Sprintf("ssl_certchainfile present with non-default value (expected none or omitted), got: %s", httpdConfig.SslCertchainfile))
		}
		if httpdConfig.SslInclude != "" && httpdConfig.SslInclude != "none" {
			errors = append(errors, fmt.Sprintf("ssl_include present with non-default value (expected none or omitted), got: %s", httpdConfig.SslInclude))
		}
		if httpdConfig.SslOcspEnable != "off" {
			errors = append(errors, fmt.Sprintf("ssl_ocsp_enable not reset to default (off), got: %s", httpdConfig.SslOcspEnable))
		}
		if httpdConfig.SslVerifyClient != "no" {
			errors = append(errors, fmt.Sprintf("ssl_verify_client not reset to default (no), got: %s", httpdConfig.SslVerifyClient))
		}
		if httpdConfig.SslVerifyDepth != 10 {
			errors = append(errors, fmt.Sprintf("ssl_verify_depth not reset to default (10), got: %d", httpdConfig.SslVerifyDepth))
		}

		if len(errors) > 0 {
			errorMsg := "HTTPD configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] HTTPD configuration successfully reset to default values")
	}
	return nil
}
