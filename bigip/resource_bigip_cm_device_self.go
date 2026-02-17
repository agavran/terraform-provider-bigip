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

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validateMirrorIP validates mirror_ip and mirror_secondary_ip fields
// Accepts: valid IP addresses, "any6", or "any"
func validateMirrorIP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "any6" || v == "any" {
		return
	}
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%q must be a valid IP address, 'any6', or 'any', got: %s", key, v))
	}
	return
}

// validateConfigsyncIP validates configsync_ip field
// Accepts: valid IP addresses or "none"
func validateConfigsyncIP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "none" {
		return
	}
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%q must be a valid IP address or 'none', got: %s", key, v))
	}
	return
}

// validateUnicastIP validates IP fields in unicast_address
// Accepts: valid IP addresses or "management-ip"
func validateUnicastIP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "management-ip" {
		return
	}
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%q must be a valid IP address or 'management-ip', got: %s", key, v))
	}
	return
}

func resourceBigipCmDeviceSelf() *schema.Resource {
	return &schema.Resource{
		Description: "Manages BIG-IP CM Device Self configuration for clustering and high availability. " +
			"This resource configures the local device's mirror IP, configsync IP, and unicast addresses. " +
			"NOTE: Only one instance of this resource should exist per BIG-IP device.",
		CreateContext: resourceBigipCmDeviceSelfCreate,
		UpdateContext: resourceBigipCmDeviceSelfUpdate,
		ReadContext:   resourceBigipCmDeviceSelfRead,
		DeleteContext: resourceBigipCmDeviceSelfDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the device to configure.",
			},
			"mirror_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "any6",
				ValidateFunc: validateMirrorIP,
				Description:  "IP address used for connection and persistence mirroring (default: 'any6')",
			},
			"mirror_secondary_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "any6",
				ValidateFunc: validateMirrorIP,
				Description:  "Secondary IP address used for connection and persistence mirroring (default: 'any6')",
			},
			"configsync_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validateConfigsyncIP,
				Description:  "IP address used for config sync (default: 'none')",
			},
			"multicast_interface": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Interface for multicast failover (default: '')",
			},
			"multicast_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "any6",
				ValidateFunc: validateMirrorIP,
				Description:  "IP address for multicast failover (default: 'any6')",
			},
			"multicast_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 65535),
				Description:  "Port for multicast failover (default: 0)",
			},
			"unicast_address": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of unicast addresses for failover communication",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"effective_ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateUnicastIP,
							Description:  "The effective IP address for unicast failover (valid IP or 'management-ip')",
						},
						"effective_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1026,
							ValidateFunc: validation.IntBetween(1, 65535),
							Description:  "The effective port for unicast failover (default: 1026)",
						},
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateUnicastIP,
							Description:  "The IP address for unicast failover (valid IP or 'management-ip')",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1026,
							ValidateFunc: validation.IntBetween(1, 65535),
							Description:  "The port for unicast failover (default: 1026)",
						},
					},
				},
			},
		},
	}
}

// buildDeviceSelfConfig constructs a DeviceSelf config from ResourceData
func buildDeviceSelfConfig(d *schema.ResourceData) *bigip.DeviceSelf {
	config := &bigip.DeviceSelf{}

	if val, ok := d.GetOk("name"); ok {
		config.Name = val.(string)
	}
	if val, ok := d.GetOk("mirror_ip"); ok {
		config.MirrorIp = val.(string)
	}
	if val, ok := d.GetOk("mirror_secondary_ip"); ok {
		config.MirrorSecondaryIp = val.(string)
	}
	if val, ok := d.GetOk("configsync_ip"); ok {
		config.ConfigsyncIp = val.(string)
	}
	if val, ok := d.GetOk("multicast_interface"); ok {
		config.MulticastInterface = val.(string)
	}
	if val, ok := d.GetOk("multicast_ip"); ok {
		config.MulticastIp = val.(string)
	}
	config.MulticastPort = d.Get("multicast_port").(int)
	if val, ok := d.GetOk("unicast_address"); ok {
		unicastList := val.([]interface{})
		config.UnicastAddress = make([]bigip.UnicastAddress, len(unicastList))
		for i, item := range unicastList {
			unicast := item.(map[string]interface{})
			ip := unicast["ip"].(string)
			port := unicast["port"].(int)
			effectiveIP := unicast["effective_ip"].(string)
			effectivePort := unicast["effective_port"].(int)
			config.UnicastAddress[i] = bigip.UnicastAddress{
				IP:            ip,
				Port:          port,
				EffectiveIP:   effectiveIP,
				EffectivePort: effectivePort,
			}
		}
	}

	return config
}

func resourceBigipCmDeviceSelfCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating CM Device Self Configuration")
	config := buildDeviceSelfConfig(d)

	if err := client.CreateDeviceSelf(config); err != nil {
		log.Printf("[ERROR] Unable to Create CM Device Self Config: %v", err)
		return diag.FromErr(err)
	}

	d.SetId(d.Get("name").(string))
	return resourceBigipCmDeviceSelfRead(ctx, d, meta)
}

func resourceBigipCmDeviceSelfUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating CM Device Self Configuration")
	config := buildDeviceSelfConfig(d)

	if err := client.ModifyDeviceSelf(config); err != nil {
		log.Printf("[ERROR] Unable to Modify CM Device Self Config: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipCmDeviceSelfRead(ctx, d, meta)
}

func resourceBigipCmDeviceSelfRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading CM Device Self Configuration for %s", name)

	obj, err := client.GetDeviceSelf(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve CM Device Self Config: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] CM Device Self Config not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", obj.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("mirror_ip", obj.MirrorIp); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("mirror_secondary_ip", obj.MirrorSecondaryIp); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("configsync_ip", obj.ConfigsyncIp); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if obj.MulticastInterface != "" {
		if err := d.Set("multicast_interface", obj.MulticastInterface); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}
	if err := d.Set("multicast_ip", obj.MulticastIp); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("multicast_port", obj.MulticastPort); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Convert unicast addresses to list of maps
	unicastAddresses := make([]map[string]interface{}, len(obj.UnicastAddress))
	for i, ua := range obj.UnicastAddress {
		unicastAddresses[i] = map[string]interface{}{
			"effective_ip":   ua.EffectiveIP,
			"effective_port": ua.EffectivePort,
			"ip":             ua.IP,
			"port":           ua.Port,
		}
	}
	if err := d.Set("unicast_address", unicastAddresses); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipCmDeviceSelfDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting CM Device Self Configuration to defaults")

	name := d.Get("name").(string)
	err := client.DeleteDeviceSelf(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Reset CM Device Self Config: %v", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
