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

func resourceBigipSysAuthLdap() *schema.Resource {
	return &schema.Resource{
		Description: "Manages BIG-IP LDAP authentication configuration. " +
			"NOTE: BIG-IP only supports one LDAP authentication configuration named 'system-auth'. " +
			"Only one instance of this resource should exist per BIG-IP device. " +
			"If the LDAP config already exists on the device, use `terraform import` instead of creating a new resource.",
		CreateContext: resourceBigipSysAuthLdapCreate,
		UpdateContext: resourceBigipSysAuthLdapUpdate,
		ReadContext:   resourceBigipSysAuthLdapRead,
		DeleteContext: resourceBigipSysAuthLdapDelete,
		CustomizeDiff: validateLdapConfigIsUnique,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Name of the LDAP configuration. Must be 'system-auth' as BIG-IP only supports one LDAP authentication configuration. " +
					"Only one instance of this resource should exist per BIG-IP device.",
				ValidateFunc: validation.StringInSlice([]string{"system-auth"}, false),
			},
			"servers": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of LDAP server addresses or hostnames",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      389,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "LDAP server port (default: 389, use 636 for LDAPS)",
			},
			"bind_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Distinguished name for binding to the LDAP server",
			},
			"bind_pw": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: "Password for binding to the LDAP server. " +
					"**Security Note:** This value is stored in Terraform state in plaintext. " +
					"Use encrypted remote state (Terraform Cloud, S3 with encryption, etc.).",
			},
			"bind_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Bind timeout in seconds (default: 30)",
			},
			"search_base_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base DN for LDAP searches",
			},
			"search_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Search timeout in seconds (default: 30)",
			},
			"login_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "samaccountname",
				Description: "LDAP attribute to use for login (default: samaccountname)",
			},
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LDAP search filter",
			},
			"group_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base DN for group searches",
			},
			"group_member_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LDAP attribute for group membership",
			},
			"check_host_attr": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Check host attribute (default: disabled)",
			},
			"check_roles_group": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Check roles group (default: disabled)",
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Idle timeout in seconds (default: 3600)",
			},
			"ignore_auth_info_unavail": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "no",
				ValidateFunc: validation.StringInSlice([]string{"yes", "no"}, false),
				Description:  "Ignore authentication info unavailable (default: no)",
			},
			"ignore_unknown_user": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Ignore unknown user (default: disabled)",
			},
			"debug": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable debug logging for LDAP authentication (default: disabled)",
			},
			"user_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User template for LDAP authentication",
			},
			"warnings": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable warning messages (default: enabled)",
			},
			"referrals": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "yes",
				ValidateFunc: validation.StringInSlice([]string{"yes", "no"}, false),
				Description:  "Follow LDAP referrals (default: yes)",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "sub",
				ValidateFunc: validation.StringInSlice([]string{"base", "one", "sub"}, false),
				Description:  "LDAP search scope (default: sub)",
			},
			"ssl": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled", "start-tls"}, false),
				Description:  "SSL mode (default: disabled)",
			},
			"ssl_ca_cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSL CA certificate file path",
			},
			"ssl_check_peer": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Verify SSL peer certificate (default: disabled)",
			},
			"ssl_ciphers": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSL cipher suite",
			},
			"ssl_client_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSL client certificate file path",
			},
			"ssl_client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSL client key file path",
			},
			"version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntInSlice([]int{2, 3}),
				Description:  "LDAP protocol version (default: 3)",
			},
		},
	}
}

// validateLdapConfigIsUnique checks if LDAP config already exists during plan phase.
func validateLdapConfigIsUnique(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() != "" {
		return nil
	}

	name := diff.Get("name").(string)
	client := meta.(*bigip.BigIP)

	existing, err := client.GetLdapConfig(name)
	if err != nil {
		return fmt.Errorf("unable to validate LDAP configuration. Cannot reach BIG-IP: %v", err)
	}

	if existing != nil {
		return fmt.Errorf("LDAP configuration '%s' already exists on BIG-IP. "+
			"BIG-IP only supports one LDAP authentication configuration. "+
			"If you have multiple 'bigip_sys_auth_ldap' resources in your configuration, remove the duplicates. "+
			"If the config exists on the device, you can use: terraform import bigip_sys_auth_ldap.<name> %s",
			name, name)
	}

	return nil
}

// buildLdapConfig constructs an LdapConfig from ResourceData
func buildLdapConfig(d *schema.ResourceData) *bigip.LdapConfig {
	config := &bigip.LdapConfig{
		Name: d.Get("name").(string),
	}

	if val, ok := d.GetOk("servers"); ok {
		servers := val.([]interface{})
		config.Servers = make([]string, len(servers))
		for i, v := range servers {
			config.Servers[i] = v.(string)
		}
	}

	if val, ok := d.GetOk("port"); ok {
		config.Port = val.(int)
	}
	if val, ok := d.GetOk("bind_dn"); ok {
		config.BindDn = val.(string)
	}
	if val, ok := d.GetOk("bind_pw"); ok {
		config.BindPw = val.(string)
	}
	if val, ok := d.GetOk("bind_timeout"); ok {
		config.BindTimeout = val.(int)
	}
	if val, ok := d.GetOk("search_base_dn"); ok {
		config.SearchBaseDn = val.(string)
	}
	if val, ok := d.GetOk("search_timeout"); ok {
		config.SearchTimeout = val.(int)
	}
	if val, ok := d.GetOk("login_attribute"); ok {
		config.LoginAttribute = val.(string)
	}
	if val, ok := d.GetOk("filter"); ok {
		config.Filter = val.(string)
	}
	if val, ok := d.GetOk("group_dn"); ok {
		config.GroupDn = val.(string)
	}
	if val, ok := d.GetOk("group_member_attribute"); ok {
		config.GroupMemberAttribute = val.(string)
	}
	if val, ok := d.GetOk("check_host_attr"); ok {
		config.CheckHostAttr = val.(string)
	}
	if val, ok := d.GetOk("check_roles_group"); ok {
		config.CheckRolesGroup = val.(string)
	}
	if val, ok := d.GetOk("idle_timeout"); ok {
		config.IdleTimeout = val.(int)
	}
	if val, ok := d.GetOk("ignore_auth_info_unavail"); ok {
		config.IgnoreAuthInfoUnavail = val.(string)
	}
	if val, ok := d.GetOk("ignore_unknown_user"); ok {
		config.IgnoreUnknownUser = val.(string)
	}
	if val, ok := d.GetOk("debug"); ok {
		config.Debug = val.(string)
	}
	if val, ok := d.GetOk("user_template"); ok {
		config.UserTemplate = val.(string)
	}
	if val, ok := d.GetOk("warnings"); ok {
		config.Warnings = val.(string)
	}
	if val, ok := d.GetOk("referrals"); ok {
		config.Referrals = val.(string)
	}
	if val, ok := d.GetOk("scope"); ok {
		config.Scope = val.(string)
	}
	if val, ok := d.GetOk("ssl"); ok {
		config.Ssl = val.(string)
	}
	if val, ok := d.GetOk("ssl_ca_cert_file"); ok {
		config.SslCaCertFile = val.(string)
	}
	if val, ok := d.GetOk("ssl_check_peer"); ok {
		config.SslCheckPeer = val.(string)
	}
	if val, ok := d.GetOk("ssl_ciphers"); ok {
		config.SslCiphers = val.(string)
	}
	if val, ok := d.GetOk("ssl_client_cert"); ok {
		config.SslClientCert = val.(string)
	}
	if val, ok := d.GetOk("ssl_client_key"); ok {
		config.SslClientKey = val.(string)
	}
	if val, ok := d.GetOk("version"); ok {
		config.Version = val.(int)
	}

	return config
}

func resourceBigipSysAuthLdapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating LDAP Authentication Config: %s", name)

	existing, err := client.GetLdapConfig(name)
	if err == nil && existing != nil {
		return diag.Errorf("LDAP configuration '%s' already exists on BIG-IP. "+
			"BIG-IP only supports one LDAP authentication configuration. "+
			"Use 'terraform import bigip_sys_auth_ldap.<resource_name> %s' to import the existing configuration instead of creating a new one.",
			name, name)
	}

	config := buildLdapConfig(d)

	if err := client.CreateLdapConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Create LDAP Config (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysAuthLdapRead(ctx, d, meta)
}

func resourceBigipSysAuthLdapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating LDAP Authentication Config: %s", name)
	config := buildLdapConfig(d)
	config.Name = name

	if err := client.ModifyLdapConfig(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify LDAP Config (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysAuthLdapRead(ctx, d, meta)
}

// setAttrLdapStringWithDiag is a helper that sets an optional string attribute and returns diagnostics.
func setAttrLdapStringWithDiag(d *schema.ResourceData, attrName string, val string) diag.Diagnostics {
	if val != "" {
		if err := d.Set(attrName, val); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

// setAttrLdapIntWithDiag is a helper that sets an optional int attribute and returns diagnostics.
func setAttrLdapIntWithDiag(d *schema.ResourceData, attrName string, val int) diag.Diagnostics {
	if val != 0 {
		if err := d.Set(attrName, val); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceBigipSysAuthLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading LDAP Authentication Config: %s", name)

	obj, err := client.GetLdapConfig(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve LDAP Config (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] LDAP Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", obj.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("servers", obj.Servers); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	diags = append(diags, setAttrLdapIntWithDiag(d, "port", obj.Port)...)
	diags = append(diags, setAttrLdapIntWithDiag(d, "bind_timeout", obj.BindTimeout)...)
	diags = append(diags, setAttrLdapIntWithDiag(d, "search_timeout", obj.SearchTimeout)...)
	diags = append(diags, setAttrLdapIntWithDiag(d, "idle_timeout", obj.IdleTimeout)...)
	diags = append(diags, setAttrLdapIntWithDiag(d, "version", obj.Version)...)

	diags = append(diags, setAttrLdapStringWithDiag(d, "bind_dn", obj.BindDn)...)
	// Don't read bind_pw back (sensitive field)
	diags = append(diags, setAttrLdapStringWithDiag(d, "search_base_dn", obj.SearchBaseDn)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "login_attribute", obj.LoginAttribute)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "filter", obj.Filter)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "group_dn", obj.GroupDn)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "group_member_attribute", obj.GroupMemberAttribute)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "check_host_attr", obj.CheckHostAttr)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "check_roles_group", obj.CheckRolesGroup)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ignore_auth_info_unavail", obj.IgnoreAuthInfoUnavail)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ignore_unknown_user", obj.IgnoreUnknownUser)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "debug", obj.Debug)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "user_template", obj.UserTemplate)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "warnings", obj.Warnings)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "referrals", obj.Referrals)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "scope", obj.Scope)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl", obj.Ssl)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl_ca_cert_file", obj.SslCaCertFile)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl_check_peer", obj.SslCheckPeer)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl_ciphers", obj.SslCiphers)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl_client_cert", obj.SslClientCert)...)
	diags = append(diags, setAttrLdapStringWithDiag(d, "ssl_client_key", obj.SslClientKey)...)

	return diags
}

func resourceBigipSysAuthLdapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting LDAP Authentication Config: %s", name)

	err := client.DeleteLdapConfig(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete LDAP Config (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
