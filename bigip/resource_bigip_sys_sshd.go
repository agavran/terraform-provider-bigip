/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// defaultAllowALL returns the default value for the allow field
func defaultAllowALL() (interface{}, error) {
	return []interface{}{"ALL"}, nil
}

// validateAllowList validates that allow list contains either "ALL" as single element,
// or valid IP addresses/networks in CIDR notation (e.g., 192.168.0.0/16 or 192.168.0.0/255.255.0.0)
func validateAllowList(val interface{}, key string) (warns []string, errs []error) {
	list, ok := val.([]interface{})
	if !ok {
		errs = append(errs, fmt.Errorf("%q must be a list", key))
		return
	}

	if len(list) == 0 {
		errs = append(errs, fmt.Errorf("%q cannot be empty", key))
		return
	}

	if len(list) == 1 {
		if str, ok := list[0].(string); ok && str == "ALL" {
			return
		}
	}

	for _, item := range list {
		if str, ok := item.(string); ok && str == "ALL" {
			errs = append(errs, fmt.Errorf("%q: when using 'ALL', it must be the only element in the list", key))
			return
		}
	}

	for i, item := range list {
		str, ok := item.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("%q[%d]: must be a string", key, i))
			continue
		}

		if strings.Contains(str, "/") {
			parts := strings.Split(str, "/")
			if len(parts) != 2 {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid format %q, expected IP/mask or IP/CIDR", key, i, str))
				continue
			}

			if net.ParseIP(parts[0]) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid IP address %q", key, i, parts[0]))
				continue
			}

			if _, _, err := net.ParseCIDR(str); err == nil {
				continue
			}

			if net.ParseIP(parts[1]) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid CIDR or netmask %q", key, i, str))
			}
		} else {
			if net.ParseIP(str) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid IP address %q", key, i, str))
			}
		}
	}

	return
}

// validateAllowListDiff validates the allow list during plan phase using CustomizeDiff
func validateAllowListDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if val, ok := diff.GetOk("allow"); ok {
		_, errs := validateAllowList(val, "allow")
		if len(errs) > 0 {
			if len(errs) == 1 {
				return errs[0]
			}
			errMsg := "Multiple validation errors for \"allow\" field:\n"
			for i, err := range errs {
				errMsg += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
			}
			return fmt.Errorf("%s", errMsg)
		}
	}
	return nil
}

func resourceBigipSysSshd() *schema.Resource {
	return &schema.Resource{
		Description: "Manages BIG-IP SSH daemon configuration. " +
			"NOTE: Only one instance of this resource should exist per BIG-IP device. " +
			"F5 Networks recommends that users of the Configuration utility exit the utility before changes are made to the system using the sshd component." +
			"This is because making changes to the system using this component causes a restart of the sshd daemon." +
			"Likewise, restarting the sshd daemon creates the necessity for a restart of the Configuration utility.",
		CreateContext: resourceBigipSysSshdCreate,
		UpdateContext: resourceBigipSysSshdUpdate,
		ReadContext:   resourceBigipSysSshdRead,
		DeleteContext: resourceBigipSysSshdDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateAllowListDiff,
		Schema: map[string]*schema.Schema{
			"allow": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				DefaultFunc: defaultAllowALL,
				Description: "IP addresses or networks (CIDR) allowed to access SSH (default: ALL)",
			},
			"banner": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable or disable SSH banner (default: disabled)",
			},
			"banner_text": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "SSH banner text to display before login",
			},
			"fips_cipher_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "FIPS cipher version - read-only, set by system based on FIPS mode",
			},
			"inactivity_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "SSH session inactivity timeout in seconds (0 = disabled)",
			},
			"include": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "Include custom SSH configuration file",
			},
			"log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "info",
				ValidateFunc: validation.StringInSlice([]string{"quiet", "fatal", "error", "info", "verbose", "debug", "debug1", "debug2", "debug3"}, false),
				Description:  "SSH daemon log level (default: info)",
			},
			"login": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable or disable SSH login (default: enabled)",
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      22,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "SSH daemon listening port (default: 22)",
			},
		},
	}
}

// buildSshdConfig constructs an SSHDConfig from ResourceData
func buildSshdConfig(d *schema.ResourceData) *bigip.SSHDConfig {
	config := &bigip.SSHDConfig{}

	if val, ok := d.GetOk("allow"); ok {
		allowList := val.([]interface{})
		allow := make([]string, len(allowList))
		for i, v := range allowList {
			allow[i] = v.(string)
		}
		config.Allow = allow
	}

	setStringIfOk(d, "banner", &config.Banner)
	setStringIfOk(d, "banner_text", &config.BannerText)
	setStringIfOk(d, "include", &config.Include)
	setStringIfOk(d, "log_level", &config.LogLevel)
	setStringIfOk(d, "login", &config.Login)

	if val, ok := d.GetOk("inactivity_timeout"); ok {
		config.InactivityTimeout = val.(int)
	}
	if val, ok := d.GetOk("port"); ok {
		config.Port = val.(int)
	}

	return config
}

func resourceBigipSysSshdCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating SSHD configuration")
	config := buildSshdConfig(d)

	if err := client.CreateSSHDConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Create SSHD configuration: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("sshd_config")
	return resourceBigipSysSshdRead(ctx, d, meta)
}

func resourceBigipSysSshdUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating SSHD configuration")
	config := buildSshdConfig(d)

	if d.HasChange("banner_text") {
		if _, ok := d.GetOk("banner_text"); !ok {
			log.Printf("[INFO] Banner text removed from config, clearing it")
			config.BannerText = "none"
		}
	}
	if d.HasChange("include") {
		if _, ok := d.GetOk("include"); !ok {
			log.Printf("[INFO] Include removed from config, clearing it")
			config.Include = "none"
		}
	}

	if err := client.ModifySSHDConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Modify SSHD configuration: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysSshdRead(ctx, d, meta)
}

func resourceBigipSysSshdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading SSHD configuration")

	obj, err := client.GetSSHDConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve SSHD configuration: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] SSHD configuration not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	setResourceData(d, "allow", obj.Allow, &diags)
	setResourceData(d, "banner", obj.Banner, &diags)

	if obj.BannerText != "" && obj.BannerText != "none" {
		setResourceData(d, "banner_text", obj.BannerText, &diags)
	}

	setResourceData(d, "fips_cipher_version", obj.FipsCipherVersion, &diags)
	setResourceData(d, "inactivity_timeout", obj.InactivityTimeout, &diags)

	if obj.Include != "" && obj.Include != "none" {
		setResourceData(d, "include", obj.Include, &diags)
	}

	setResourceData(d, "log_level", obj.LogLevel, &diags)
	setResourceData(d, "login", obj.Login, &diags)
	setResourceData(d, "port", obj.Port, &diags)

	return diags
}

func resourceBigipSysSshdDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting SSHD configuration to defaults")

	err := client.DeleteSSHDConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset SSHD configuration (%v)", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
