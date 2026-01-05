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

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validateSnmpUserPasswords validates protocols and passwords based on security_level
func validateSnmpUserPasswords(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	securityLevel, secOk := diff.GetOk("security_level")
	authProtocol, authOk := diff.GetOk("auth_protocol")
	authPassword, authPassOk := diff.GetOk("auth_password")
	privacyProtocol, privOk := diff.GetOk("privacy_protocol")
	privacyPassword, privPassOk := diff.GetOk("privacy_password")

	var errors []string

	if !secOk {
		return fmt.Errorf("security_level is required but not set")
	}

	secLevel := securityLevel.(string)
	authProto := ""
	privProto := ""

	if authOk {
		authProto = authProtocol.(string)
	}
	if privOk {
		privProto = privacyProtocol.(string)
	}

	if authOk && authProto == "none" {
		if authPassOk && authPassword.(string) != "" {
			errors = append(errors, "auth_password cannot be set when auth_protocol is 'none'")
		}
	}
	if privOk && privProto == "none" {
		if privPassOk && privacyPassword.(string) != "" {
			errors = append(errors, "privacy_password cannot be set when privacy_protocol is 'none'")
		}
	}

	switch secLevel {
	case "no-auth-no-privacy":
		if authOk && authProto != "none" {
			errors = append(errors, "auth_protocol must be 'none' when security_level is 'no-auth-no-privacy'")
		}
		if privOk && privProto != "none" {
			errors = append(errors, "privacy_protocol must be 'none' when security_level is 'no-auth-no-privacy'")
		}
	case "auth-no-privacy":
		if authOk && authProto == "none" {
			errors = append(errors, "auth_protocol cannot be 'none' when security_level is 'auth-no-privacy'")
		}
		if privOk && privProto != "none" {
			errors = append(errors, "privacy_protocol must be 'none' when security_level is 'auth-no-privacy'")
		}
		if authOk && authProto != "none" {
			if !authPassOk || authPassword.(string) == "" {
				errors = append(errors, "auth_password is required when auth_protocol is not 'none'")
			}
		}
	case "auth-privacy":
		if authOk && authProto == "none" {
			errors = append(errors, "auth_protocol cannot be 'none' when security_level is 'auth-privacy'")
		}
		if privOk && privProto == "none" {
			errors = append(errors, "privacy_protocol cannot be 'none' when security_level is 'auth-privacy'")
		}
		if authOk && authProto != "none" {
			if !authPassOk || authPassword.(string) == "" {
				errors = append(errors, "auth_password is required when auth_protocol is not 'none'")
			}
		}
		if privOk && privProto != "none" {
			if !privPassOk || privacyPassword.(string) == "" {
				errors = append(errors, "privacy_password is required when privacy_protocol is not 'none'")
			}
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return fmt.Errorf("%s", errors[0])
		}
		errMsg := "Multiple validation errors:\n"
		for i, err := range errors {
			errMsg += fmt.Sprintf("  %d. %s\n", i+1, err)
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

func resourceBigipSysSnmpUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysSnmpUserCreate,
		UpdateContext: resourceBigipSysSnmpUserUpdate,
		ReadContext:   resourceBigipSysSnmpUserRead,
		DeleteContext: resourceBigipSysSnmpUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateSnmpUserPasswords,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5Name,
				Description:  "Name of the SNMP user",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SNMPv3 username",
			},
			"auth_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "sha",
				ValidateFunc: validation.StringInSlice([]string{"md5", "sha", "sha256", "sha512", "none"}, false),
				Description:  "Authentication protocol (md5, sha, sha256, sha512, none)",
			},
			"auth_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 255),
				Description:  "Authentication password (required unless auth_protocol is 'none')",
			},
			"privacy_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "aes",
				ValidateFunc: validation.StringInSlice([]string{"aes", "aes192", "aes256", "des", "none"}, false),
				Description:  "Privacy/encryption protocol (aes, aes192, aes256, des, none)",
			},
			"privacy_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 255),
				Description:  "Privacy/encryption password (required unless privacy_protocol is 'none')",
			},
			"access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ro",
				ValidateFunc: validation.StringInSlice([]string{"ro", "rw", "roa", "rwa"}, false),
				Description:  "Access level (ro=read-only, rw=read-write, roa=read-only-all, rwa=read-write-all, default: ro)",
			},
			"security_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "auth-privacy",
				ValidateFunc: validation.StringInSlice([]string{"no-auth-no-privacy", "auth-no-privacy", "auth-privacy"}, false),
				Description:  "Security level (no-auth-no-privacy, auth-no-privacy, auth-privacy, default: auth-privacy)",
			},
			"oid_subset": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "OID subset for restricting SNMP access",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the SNMP user",
			},
		},
	}
}

// buildSnmpUserConfig constructs an SnmpUser config from ResourceData
func buildSnmpUserConfig(d *schema.ResourceData) *bigip.SnmpUser {
	config := &bigip.SnmpUser{
		Name:            d.Get("name").(string),
		Username:        d.Get("username").(string),
		AuthProtocol:    d.Get("auth_protocol").(string),
		PrivacyProtocol: d.Get("privacy_protocol").(string),
	}

	if val, ok := d.GetOk("auth_password"); ok {
		config.AuthPassword = val.(string)
	}
	if val, ok := d.GetOk("privacy_password"); ok {
		config.PrivacyPassword = val.(string)
	}
	if val, ok := d.GetOk("access"); ok {
		config.Access = val.(string)
	}
	if val, ok := d.GetOk("security_level"); ok {
		config.SecurityLevel = val.(string)
	}
	if val, ok := d.GetOk("oid_subset"); ok {
		config.OidSubset = val.(string)
	}
	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}

	return config
}

func resourceBigipSysSnmpUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating SNMP User: %s", name)
	config := buildSnmpUserConfig(d)

	if err := client.CreateSnmpUser(config); err != nil {
		log.Printf("[ERROR] Unable to Create SNMP User (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysSnmpUserRead(ctx, d, meta)
}

func resourceBigipSysSnmpUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating SNMP User: %s", name)
	config := buildSnmpUserConfig(d)
	config.Name = name // Ensure we use the ID for updates

	if err := client.ModifySnmpUser(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify SNMP User (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysSnmpUserRead(ctx, d, meta)
}

func resourceBigipSysSnmpUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading SNMP User: %s", name)

	obj, err := client.GetSnmpUser(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve SNMP User (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] SNMP User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Collect all errors instead of ignoring them
	var diags diag.Diagnostics

	if err := d.Set("name", obj.FullPath); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("username", obj.Username); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("auth_protocol", obj.AuthProtocol); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	// Don't read passwords back (sensitive fields)
	if err := d.Set("privacy_protocol", obj.PrivacyProtocol); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("access", obj.Access); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("security_level", obj.SecurityLevel); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("oid_subset", obj.OidSubset); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", obj.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysSnmpUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting SNMP User: %s", name)

	err := client.DeleteSnmpUser(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete SNMP User (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
