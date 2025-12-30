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

func resourceBigipSysGlobalSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysGlobalSettingsCreate,
		UpdateContext: resourceBigipSysGlobalSettingsUpdate,
		ReadContext:   resourceBigipSysGlobalSettingsRead,
		DeleteContext: resourceBigipSysGlobalSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"gui_security_banner": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable or disable the GUI security banner (default: enabled)",
			},
			"gui_security_banner_text": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Welcome to the BIG-IP Configuration Utility.\n\nLog in with your username and password using the fields on the left.",
				Description: "Text to display in the GUI security banner",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "bigip1",
				Description: "System hostname",
			},
		},
	}
}

// buildGlobalSettingsConfig constructs a GlobalSettings config from ResourceData
func buildGlobalSettingsConfig(d *schema.ResourceData) *bigip.GlobalSettings {
	config := &bigip.GlobalSettings{}

	if val, ok := d.GetOk("gui_security_banner"); ok {
		config.GuiSecurityBanner = val.(string)
	}
	if val, ok := d.GetOk("gui_security_banner_text"); ok {
		config.GuiSecurityBannerText = val.(string)
	}
	if val, ok := d.GetOk("hostname"); ok {
		config.Hostname = val.(string)
	}

	return config
}

func resourceBigipSysGlobalSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating Global Settings configuration")
	config := buildGlobalSettingsConfig(d)

	if err := client.CreateGlobalSettings(config); err != nil {
		log.Printf("[ERROR] Unable to Create Global Settings: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("global_settings")
	return resourceBigipSysGlobalSettingsRead(ctx, d, meta)
}

func resourceBigipSysGlobalSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating Global Settings configuration")
	config := buildGlobalSettingsConfig(d)

	if err := client.ModifyGlobalSettings(config); err != nil {
		log.Printf("[ERROR] Unable to Modify Global Settings: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysGlobalSettingsRead(ctx, d, meta)
}

func resourceBigipSysGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading Global Settings configuration")

	obj, err := client.GetGlobalSettings()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Global Settings: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Global Settings not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("gui_security_banner", obj.GuiSecurityBanner); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("gui_security_banner_text", obj.GuiSecurityBannerText); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("hostname", obj.Hostname); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysGlobalSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting Global Settings to defaults")

	err := client.DeleteGlobalSettings()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset Global Settings (%v)", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
