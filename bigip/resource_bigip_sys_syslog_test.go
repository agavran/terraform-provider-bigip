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

var TEST_SYSLOG_NAME = "syslog_config"
var TEST_SYSLOG_RESOURCE_NAME = "bigip_sys_syslog.test-syslog"
var TEST_SYSLOG_RESOURCE = `
resource "bigip_sys_syslog" "test-syslog" {
	console_log     = "enabled"
	iso_date        = "enabled"
	auth_priv_from  = "info"
	daemon_from     = "info"
	kern_from       = "notice"

	remote_servers {
		name        = "/Common/syslog-server-1"
		host        = "10.1.1.100"
		remote_port = 514
		local_ip    = "none"
		description = "none"
	}

	remote_servers {
		name        = "/Common/syslog-server-2"
		host        = "syslog.example.com"
		remote_port = 1514
	}
}
`
var TEST_SYSLOG_RESOURCE_UPDATE = `
resource "bigip_sys_syslog" "test-syslog" {
	console_log     = "disabled"
	iso_date        = "disabled"
	auth_priv_from  = "warning"
	daemon_from     = "warning"
	kern_from       = "err"
	description     = "Testing include statement"
	include         = <<-EOT
		filter f_complete {
		  facility(auth,authpriv);
		};
		destination d_syslog_server {
		  udp(\"10.1.10.251\" port (1514));
		};
		log {
		  source(s_syslog_pipe);
		  filter(f_complete);
		  destination(d_syslog_server);
		};
	EOT
	
}
`
var TEST_SYSLOG_RESOURCE_UPDATE_2 = `
resource "bigip_sys_syslog" "test-syslog" {
	console_log     = "disabled"
	iso_date        = "disabled"
	auth_priv_from  = "warning"
	daemon_from     = "warning"
	kern_from       = "err"
	description     = "none"
	include         = "none"

	remote_servers {
		name        = "/Common/syslog-server-updated"
		host        = "192.168.1.200"
		remote_port = 514
		description = "Updated syslog server"
	}
}
`

func TestAccBigipSysSyslog_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYSLOG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSyslogExists(true),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "console_log", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "iso_date", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "auth_priv_from", "info"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "daemon_from", "info"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "kern_from", "notice"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.name", "/Common/syslog-server-1"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.host", "10.1.1.100"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.remote_port", "514"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.description", "Primary syslog server"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.1.name", "/Common/syslog-server-2"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.1.host", "syslog.example.com"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.1.remote_port", "1514"),
				),
			},
		},
	})
}

func TestAccBigipSysSyslog_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSyslogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYSLOG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "console_log", "enabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.#", "2"),
				),
			},
			{
				Config: TEST_SYSLOG_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "console_log", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "iso_date", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "auth_priv_from", "warning"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "daemon_from", "warning"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "kern_from", "err"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "description", "Testing include statement"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.#", "0"),
				),
			},
			{
				Config: TEST_SYSLOG_RESOURCE_UPDATE_2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "console_log", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "iso_date", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "auth_priv_from", "warning"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "daemon_from", "warning"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "kern_from", "err"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "description", ""),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.name", "/Common/syslog-server-updated"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.host", "192.168.1.200"),
					resource.TestCheckResourceAttr("bigip_sys_syslog.test-syslog", "remote_servers.0.description", "Updated syslog server"),
				),
			},
		},
	})
}

func TestAccBigipSysSyslog_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSyslogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SYSLOG_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSyslogExists(true),
				),
			},
			{
				ResourceName:      TEST_SYSLOG_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SYSLOG_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckSyslogExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetSyslogConfig()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("Syslog configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("Syslog configuration still exists.")
		}
		return nil
	}
}

func testCheckSyslogDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_syslog" {
			continue
		}

		syslogConfig, err := client.GetSyslogConfig()
		if err != nil {
			return err
		}

		var errors []string

		if syslogConfig.AuthPrivFrom != "notice" {
			errors = append(errors, fmt.Sprintf("auth_priv_from not reset to default (notice), got: %s", syslogConfig.AuthPrivFrom))
		}
		if syslogConfig.AuthPrivTo != "emerg" {
			errors = append(errors, fmt.Sprintf("auth_priv_to not reset to default (emerg), got: %s", syslogConfig.AuthPrivTo))
		}
		if syslogConfig.ClusteredHostSlot != "enabled" {
			errors = append(errors, fmt.Sprintf("clustered_host_slot not reset to default (enabled), got: %s", syslogConfig.ClusteredHostSlot))
		}
		if syslogConfig.ClusteredMessageSlot != "disabled" {
			errors = append(errors, fmt.Sprintf("clustered_message_slot not reset to default (disabled), got: %s", syslogConfig.ClusteredMessageSlot))
		}
		if syslogConfig.ConsoleLog != "enabled" {
			errors = append(errors, fmt.Sprintf("console_log not reset to default (enabled), got: %s", syslogConfig.ConsoleLog))
		}
		if syslogConfig.CronFrom != "warning" {
			errors = append(errors, fmt.Sprintf("cron_from not reset to default (warning), got: %s", syslogConfig.CronFrom))
		}
		if syslogConfig.CronTo != "emerg" {
			errors = append(errors, fmt.Sprintf("cron_to not reset to default (emerg), got: %s", syslogConfig.CronTo))
		}
		if syslogConfig.DaemonFrom != "notice" {
			errors = append(errors, fmt.Sprintf("daemon_from not reset to default (notice), got: %s", syslogConfig.DaemonFrom))
		}
		if syslogConfig.DaemonTo != "emerg" {
			errors = append(errors, fmt.Sprintf("daemon_to not reset to default (emerg), got: %s", syslogConfig.DaemonTo))
		}
		// Description and Include should not be returned if set to default (none), but if present they should be "none" or empty
		if syslogConfig.Description != "" && syslogConfig.Description != "none" {
			errors = append(errors, fmt.Sprintf("description present with non-default value (expected none or omitted), got: %s", syslogConfig.Description))
		}
		if syslogConfig.Include != "" && syslogConfig.Include != "none" {
			errors = append(errors, fmt.Sprintf("include present with non-default value (expected none or omitted), got: %s", syslogConfig.Include))
		}
		if syslogConfig.IsoDate != "disabled" {
			errors = append(errors, fmt.Sprintf("iso_date not reset to default (disabled), got: %s", syslogConfig.IsoDate))
		}
		if syslogConfig.KernFrom != "debug" {
			errors = append(errors, fmt.Sprintf("kern_from not reset to default (debug), got: %s", syslogConfig.KernFrom))
		}
		if syslogConfig.KernTo != "emerg" {
			errors = append(errors, fmt.Sprintf("kern_to not reset to default (emerg), got: %s", syslogConfig.KernTo))
		}
		if syslogConfig.Local6From != "notice" {
			errors = append(errors, fmt.Sprintf("local6_from not reset to default (notice), got: %s", syslogConfig.Local6From))
		}
		if syslogConfig.Local6To != "emerg" {
			errors = append(errors, fmt.Sprintf("local6_to not reset to default (emerg), got: %s", syslogConfig.Local6To))
		}
		if syslogConfig.MailFrom != "notice" {
			errors = append(errors, fmt.Sprintf("mail_from not reset to default (notice), got: %s", syslogConfig.MailFrom))
		}
		if syslogConfig.MailTo != "emerg" {
			errors = append(errors, fmt.Sprintf("mail_to not reset to default (emerg), got: %s", syslogConfig.MailTo))
		}
		if syslogConfig.MessagesFrom != "notice" {
			errors = append(errors, fmt.Sprintf("messages_from not reset to default (notice), got: %s", syslogConfig.MessagesFrom))
		}
		if syslogConfig.MessagesTo != "warning" {
			errors = append(errors, fmt.Sprintf("messages_to not reset to default (warning), got: %s", syslogConfig.MessagesTo))
		}
		if syslogConfig.UserLogFrom != "notice" {
			errors = append(errors, fmt.Sprintf("user_log_from not reset to default (notice), got: %s", syslogConfig.UserLogFrom))
		}
		if syslogConfig.UserLogTo != "emerg" {
			errors = append(errors, fmt.Sprintf("user_log_to not reset to default (emerg), got: %s", syslogConfig.UserLogTo))
		}
		if len(syslogConfig.RemoteServers) != 0 {
			errors = append(errors, fmt.Sprintf("remote_servers not reset to default (empty), got %d servers", len(syslogConfig.RemoteServers)))
		}

		if len(errors) > 0 {
			errorMsg := "Syslog configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] Syslog configuration successfully reset to default values")
	}
	return nil
}
