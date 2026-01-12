/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"context"
	"log"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// suppressNoneDiff is a DiffSuppressFunc that treats empty string and "none" as equivalent
func suppressNoneDiff(k, old, new string, d *schema.ResourceData) bool {
	return (old == "" && new == "none") || (old == "none" && new == "")
}

// validateIPOrNone validates that the value is either a valid IP address or the string "none"
func validateIPOrNone(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "none" {
		return
	}
	return validation.IsIPAddress(val, key)
}

// setStringIfOk is a helper function to set a string field from ResourceData if present
func setStringIfOk(d *schema.ResourceData, key string, target *string) {
	if val, ok := d.GetOk(key); ok {
		*target = val.(string)
	}
}

// setResourceData is a helper function to set a value in ResourceData and collect any errors
func setResourceData(d *schema.ResourceData, key string, value interface{}, diags *diag.Diagnostics) {
	if err := d.Set(key, value); err != nil {
		*diags = append(*diags, diag.FromErr(err)...)
	}
}

func resourceBigipSysSyslog() *schema.Resource {
	return &schema.Resource{
		Description: "Manages BIG-IP Syslog configuration. " +
			"NOTE: Only one instance of this resource should exist per BIG-IP device. " +
			"Multiple resources will reference the same underlying configuration and overwrite each other.",
		CreateContext: resourceBigipSysSyslogCreate,
		UpdateContext: resourceBigipSysSyslogUpdate,
		ReadContext:   resourceBigipSysSyslogRead,
		DeleteContext: resourceBigipSysSyslogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"auth_priv_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for auth-priv facility (default: notice)",
			},
			"auth_priv_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for auth-priv facility (default: emerg)",
			},
			"clustered_host_slot": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable clustered host slot logging (default: enabled)",
			},
			"clustered_message_slot": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable clustered message slot logging (default: disabled)",
			},
			"console_log": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable console logging (default: enabled)",
			},
			"cron_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "warning",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for cron facility (default: warning)",
			},
			"cron_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for cron facility (default: emerg)",
			},
			"daemon_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for daemon facility (default: notice)",
			},
			"daemon_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for daemon facility (default: emerg)",
			},
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "Description of the syslog configuration",
			},
			"include": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "Include custom syslog configuration file",
			},
			"iso_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Use ISO 8601 date format in logs (default: disabled)",
			},
			"kern_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "debug",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for kernel facility (default: debug)",
			},
			"kern_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for kernel facility (default: emerg)",
			},
			"local6_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for local6 facility (default: notice)",
			},
			"local6_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for local6 facility (default: emerg)",
			},
			"mail_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for mail facility (default: notice)",
			},
			"mail_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for mail facility (default: emerg)",
			},
			"messages_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for messages facility (default: notice)",
			},
			"messages_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "warning",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for messages facility (default: warning)",
			},
			"user_log_from": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "notice",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Minimum severity level for user log facility (default: notice)",
			},
			"user_log_to": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "emerg",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "notice", "warning", "err", "crit", "alert", "emerg"}, false),
				Description:  "Maximum severity level for user log facility (default: emerg)",
			},
			"remote_servers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the remote syslog server",
						},
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Hostname or IP address of the remote syslog server",
						},
						"remote_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      514,
							ValidateFunc: validation.IntBetween(1, 65535),
							Description:  "Remote syslog server port (default: 514)",
						},
						"local_ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validateIPOrNone,
							Description:  "Local IP address to use for sending syslog messages",
						},
						"description": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressNoneDiff,
							Description:      "Description of the remote syslog server",
						},
					},
				},
				Description: "List of remote syslog servers",
			},
		},
	}
}

// buildSyslogConfig constructs a SyslogConfig from ResourceData
func buildSyslogConfig(d *schema.ResourceData) *bigip.SyslogConfig {
	config := &bigip.SyslogConfig{}

	setStringIfOk(d, "auth_priv_from", &config.AuthPrivFrom)
	setStringIfOk(d, "auth_priv_to", &config.AuthPrivTo)
	setStringIfOk(d, "clustered_host_slot", &config.ClusteredHostSlot)
	setStringIfOk(d, "clustered_message_slot", &config.ClusteredMessageSlot)
	setStringIfOk(d, "console_log", &config.ConsoleLog)
	setStringIfOk(d, "cron_from", &config.CronFrom)
	setStringIfOk(d, "cron_to", &config.CronTo)
	setStringIfOk(d, "daemon_from", &config.DaemonFrom)
	setStringIfOk(d, "daemon_to", &config.DaemonTo)
	setStringIfOk(d, "description", &config.Description)
	setStringIfOk(d, "include", &config.Include)
	setStringIfOk(d, "iso_date", &config.IsoDate)
	setStringIfOk(d, "kern_from", &config.KernFrom)
	setStringIfOk(d, "kern_to", &config.KernTo)
	setStringIfOk(d, "local6_from", &config.Local6From)
	setStringIfOk(d, "local6_to", &config.Local6To)
	setStringIfOk(d, "mail_from", &config.MailFrom)
	setStringIfOk(d, "mail_to", &config.MailTo)
	setStringIfOk(d, "messages_from", &config.MessagesFrom)
	setStringIfOk(d, "messages_to", &config.MessagesTo)
	setStringIfOk(d, "user_log_from", &config.UserLogFrom)
	setStringIfOk(d, "user_log_to", &config.UserLogTo)

	if val, ok := d.GetOk("remote_servers"); ok {
		remoteServersList := val.([]interface{})
		remoteServers := make([]bigip.SyslogRemoteServer, len(remoteServersList))
		for i, serverInterface := range remoteServersList {
			server := serverInterface.(map[string]interface{})
			remoteServer := bigip.SyslogRemoteServer{
				Name: server["name"].(string),
				Host: server["host"].(string),
			}
			if port, ok := server["remote_port"].(int); ok {
				remoteServer.RemotePort = port
			}
			if localIp, ok := server["local_ip"].(string); ok {
				remoteServer.LocalIp = localIp
			}
			if description, ok := server["description"].(string); ok && description != "" {
				remoteServer.Description = description
			} else {
				remoteServer.Description = "none"
			}
			remoteServers[i] = remoteServer
		}
		config.RemoteServers = remoteServers
	}

	return config
}

func resourceBigipSysSyslogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating Syslog configuration")
	config := buildSyslogConfig(d)

	if err := client.CreateSyslogConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Create Syslog configuration: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("syslog_config")
	return resourceBigipSysSyslogRead(ctx, d, meta)
}

func resourceBigipSysSyslogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating Syslog configuration")
	config := buildSyslogConfig(d)

	if d.HasChange("description") {
		if _, ok := d.GetOk("description"); !ok {
			log.Printf("[INFO] Description removed from config, clearing it")
			config.Description = "none"
		}
	}
	if d.HasChange("include") {
		if _, ok := d.GetOk("include"); !ok {
			log.Printf("[INFO] Include removed from config, clearing it")
			config.Include = "none"
		}
	}
	if d.HasChange("remote_servers") {
		if _, ok := d.GetOk("remote_servers"); !ok {
			log.Printf("[INFO] Remote servers removed from config, clearing them")
			config.RemoteServers = []bigip.SyslogRemoteServer{}
		}
	}

	if err := client.ModifySyslogConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Modify Syslog configuration: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysSyslogRead(ctx, d, meta)
}

func resourceBigipSysSyslogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading Syslog configuration")

	obj, err := client.GetSyslogConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Syslog configuration: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Syslog configuration not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	setResourceData(d, "auth_priv_from", obj.AuthPrivFrom, &diags)
	setResourceData(d, "auth_priv_to", obj.AuthPrivTo, &diags)
	setResourceData(d, "clustered_host_slot", obj.ClusteredHostSlot, &diags)
	setResourceData(d, "clustered_message_slot", obj.ClusteredMessageSlot, &diags)
	setResourceData(d, "console_log", obj.ConsoleLog, &diags)
	setResourceData(d, "cron_from", obj.CronFrom, &diags)
	setResourceData(d, "cron_to", obj.CronTo, &diags)
	setResourceData(d, "daemon_from", obj.DaemonFrom, &diags)
	setResourceData(d, "daemon_to", obj.DaemonTo, &diags)
	setResourceData(d, "description", obj.Description, &diags)
	setResourceData(d, "include", obj.Include, &diags)
	setResourceData(d, "iso_date", obj.IsoDate, &diags)
	setResourceData(d, "kern_from", obj.KernFrom, &diags)
	setResourceData(d, "kern_to", obj.KernTo, &diags)
	setResourceData(d, "local6_from", obj.Local6From, &diags)
	setResourceData(d, "local6_to", obj.Local6To, &diags)
	setResourceData(d, "mail_from", obj.MailFrom, &diags)
	setResourceData(d, "mail_to", obj.MailTo, &diags)
	setResourceData(d, "messages_from", obj.MessagesFrom, &diags)
	setResourceData(d, "messages_to", obj.MessagesTo, &diags)
	setResourceData(d, "user_log_from", obj.UserLogFrom, &diags)
	setResourceData(d, "user_log_to", obj.UserLogTo, &diags)

	remoteServers := make([]map[string]interface{}, len(obj.RemoteServers))
	for i, server := range obj.RemoteServers {
		remoteServer := make(map[string]interface{})
		remoteServer["name"] = server.Name
		remoteServer["host"] = server.Host
		remoteServer["remote_port"] = server.RemotePort
		remoteServer["local_ip"] = server.LocalIp

		if server.Description != "" && server.Description != "none" {
			remoteServer["description"] = server.Description
		}
		remoteServers[i] = remoteServer
	}
	setResourceData(d, "remote_servers", remoteServers, &diags)

	return diags
}

func resourceBigipSysSyslogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting Syslog configuration to defaults")

	err := client.DeleteSyslogConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset Syslog configuration (%v)", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
