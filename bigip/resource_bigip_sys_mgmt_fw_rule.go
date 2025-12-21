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
	"strconv"
	"strings"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validateIPProtocol validates IP protocol - accepts protocol name (0 - 147) or number (148 - 255)
func validateIPProtocol(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)

	if num, err := strconv.Atoi(v); err == nil {
		if num >= 148 && num <= 255 {
			return
		}
		if num >= 0 && num <= 147 {
			errs = append(errs, fmt.Errorf("%q: protocols from range 0 - 147 must be specified by name (e.g., 'tcp', 'udp', 'icmp'). Only other protocols in range 148 - 255 can be set by using numeric values. Provided value: %d", key, num))
			return
		}
		errs = append(errs, fmt.Errorf("%q protocol number must be between 148 and 255, got: %d", key, num))
		return
	}

	validProtocols := []string{
		"3pc", "a/n", "ah", "any", "argus", "aris", "ax.25", "bbn-rcc", "bna", "br-sat-mon",
		"cbt", "cftp", "chaos", "compaq-peer", "cphb", "cpnx", "crdup", "crtp", "dccp", "dcn",
		"ddp", "ddx", "dgp", "dsr", "egp", "eigrp", "emcon", "encap", "esp", "etherip",
		"fc", "fire", "ggp", "gmtp", "gre", "hip", "hmp", "hopopt", "i-nlsp", "iatp",
		"icmp", "idpr", "idpr-cmtp", "idrp", "ifmp", "igmp", "igp", "il", "ip", "ipcomp",
		"ipcv", "ipip", "iplt", "ippc", "ipv4", "ipv6", "ipv6-auth", "ipv6-crypt", "ipv6-frag",
		"ipv6-icmp", "ipv6-nonxt", "ipv6-opts", "ipv6-route", "ipx-in-ip", "irtp", "isis", "iso-ip",
		"iso-tp4", "kryptolan", "l2tp", "larp", "leaf-1", "leaf-2", "manet", "merit-inp", "mfe-nsp",
		"micp", "mobile", "mobility-header", "mpls-in-ip", "mtp", "mux", "narp", "netblt", "nsfnet-igp",
		"nvp", "ospf", "pgm", "pim", "pipe", "pnni", "prm", "ptp", "pup", "pvp",
		"qnx", "rdp", "rohc", "rsvp", "rsvp-e2e-ignore", "rvd", "sat-expak", "sat-mon", "scc-sp", "scps",
		"sctp", "sdrp", "secure-vmtp", "shim6", "skip", "sm", "smp", "snp", "sprite-rpc", "sps",
		"srp", "sscopmce", "st", "stp", "sun-nd", "swipe", "tcf", "tcp", "tlsp", "tp++",
		"trunk-1", "trunk-2", "ttp", "udp", "udplite", "uti", "vines", "visa", "vmtp", "vrrp",
		"wb-expak", "wb-mon", "wesp", "wsn", "xnet", "xns-idp", "xtp",
	}

	for _, valid := range validProtocols {
		if v == valid {
			return
		}
	}

	errs = append(errs, fmt.Errorf("%q must be a valid protocol name (for protocols 0-147) or number (148-255 for custom protocols), got: %s", key, v))
	return
}

// validateICMPCode validates ICMP format "type:code" where both are integers
func validateICMPCode(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)

	// Special case: "255" means "any type, any code"
	if v == "255" {
		return
	}

	parts := strings.Split(v, ":")
	if len(parts) != 2 {
		errs = append(errs, fmt.Errorf("%q must be in format 'type:code' (e.g., '8:0') or '255' for any, got: %s", key, v))
		return
	}
	for i, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			errs = append(errs, fmt.Errorf("%q part %d must be an integer, got: %s", key, i, part))
		}
	}
	return
}

// validateICMPProtocol ensures ICMP attributes are only used with ICMP protocols
func validateICMPProtocol(diff *schema.ResourceDiff) error {
	icmpData, ok := diff.GetOk("icmp")
	if !ok {
		return nil
	}

	icmpSet := icmpData.(*schema.Set)
	if icmpSet.Len() > 0 {
		ipProto, _ := diff.Get("ip_protocol").(string)
		if ipProto != "icmp" && ipProto != "ipv6-icmp" {
			return fmt.Errorf("ip_protocol must be set to 'icmp' or 'ipv6-icmp' when icmp attribute is enabled")
		}
	}

	return nil
}

// validatePortsForProtocol validates if ports are specified only for port-based protocols
func validatePortsForProtocol(diff *schema.ResourceDiff, ipProto string, blockName string) error {
	allowPorts := ipProto == "tcp" || ipProto == "udp" || ipProto == "sctp"

	blockRaw, ok := diff.GetOk(blockName)
	if !ok || blockRaw == nil {
		return nil
	}

	blocks := blockRaw.([]interface{})
	if len(blocks) == 0 {
		return nil
	}

	if blocks[0] == nil {
		return nil
	}

	block := blocks[0].(map[string]interface{})
	hasPorts := false

	if portSetRaw, ok := block["ports"]; ok {
		portSet := portSetRaw.(*schema.Set)
		if portSet.Len() > 0 {
			hasPorts = true
		}
	}
	if portListsRaw, ok := block["port_lists"]; ok && portListsRaw != nil {
		portLists := portListsRaw.([]interface{})
		if len(portLists) > 0 {
			hasPorts = true
		}
	}

	if hasPorts && !allowPorts {
		return fmt.Errorf("%s.ports and %s.port_lists are only allowed when ip_protocol is tcp, udp, or sctp (current: %s)", blockName, blockName, ipProto)
	}

	return nil
}

// validateUUID ensures UUID cannot be changed once set
func validateUUID(diff *schema.ResourceDiff) error {

	if diff.Id() == "" {
		return nil
	}

	if !diff.HasChange("uuid") {
		return nil
	}

	old, new := diff.GetChange("uuid")
	oldUUID := old.(string)
	newUUID := new.(string)

	if oldUUID == "" {
		return nil
	}

	if newUUID == "auto-generate" && oldUUID != "" && oldUUID != "auto-generate" {
		return nil
	}

	return fmt.Errorf(" attribute UUID cannot be modified once set. "+
		"Current UUID: %s, attempted new UUID: %s", oldUUID, newUUID)
}

// validateMgmtFwRuleConfig validates cross-field dependencies for management firewall rules
func validateMgmtFwRuleConfig(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if err := validateICMPProtocol(diff); err != nil {
		return err
	}

	ipProto, _ := diff.Get("ip_protocol").(string)
	if err := validatePortsForProtocol(diff, ipProto, "source"); err != nil {
		return err
	}
	if err := validatePortsForProtocol(diff, ipProto, "destination"); err != nil {
		return err
	}

	if err := validateUUID(diff); err != nil {
		return err
	}

	return nil
}

// suppressUUIDDiff suppresses diff when config has "auto-generate" and state has actual UUID
func suppressUUIDDiff(k, old, new string, d *schema.ResourceData) bool {
	if new == "auto-generate" && old != "" && old != "auto-generate" {
		return true
	}
	return false
}

func resourceBigipSysMgmtFwRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysMgmtFwRuleCreate,
		UpdateContext: resourceBigipSysMgmtFwRuleUpdate,
		ReadContext:   resourceBigipSysMgmtFwRuleRead,
		DeleteContext: resourceBigipSysMgmtFwRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateMgmtFwRuleConfig,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the management firewall rule",
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Management firewall rule action. Needs to be accept, drop or reject.",
				ValidateFunc: validation.StringInSlice([]string{"accept", "drop", "reject"}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the rule",
			},
			"ip_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIPProtocol,
				Description:  "IP protocol to which the rule applies (protocol name like tcp/udp/icmp for 0 - 147, or other protocol number 148 - 255)",
			},
			"uuid": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressUUIDDiff,
				Description:      "Unique identifier (assign an explict uuid based on RFC-4122 or enter auto-generate for BIGIP to create one)",
			},
			"log": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "no",
				ValidateFunc: validation.StringInSlice([]string{"yes", "no"}, false),
				Description:  "Enable rule logging (default: no)",
			},
			"place_after": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"place_after", "place_before"},
				Description:  "Rule placement in the list",
			},
			"place_before": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"place_after", "place_before"},
				Description:  "Rule placement in the list",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Firewall rule status (default: enabled)",
			},
			"schedule": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Schedule name for time-based rule activation",
			},
			"source": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Management firewall rule source data (empty block represents 'any source')",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_lists": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Address lists used for source matching",
						},
						"addresses": {
							Type:        schema.TypeSet,
							Description: "Addresses and address ranges used for source matching",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"port_lists": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Port lists used for source matching.",
						},
						"ports": {
							Type:        schema.TypeSet,
							Description: "Ports and port ranges used for source matching.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"destination": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Rule destination data (empty block represents 'any destination')",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_lists": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Address lists used for destination matching",
						},
						"addresses": {
							Type:        schema.TypeSet,
							Description: "Addresses and address ranges used for destination matching",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"port_lists": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Port lists used for destination matching",
						},
						"ports": {
							Type:        schema.TypeSet,
							Description: "Ports and port ranges used for source matching",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"icmp": {
				Type:        schema.TypeSet,
				Description: "ICMP message types and codes (format: 'type:code', e.g., '3:1' for destination unreachable, or '255' for any type/code)",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateICMPCode,
						},
					},
				},
			},
		},
	}
}

// unflattenFwRuleIpPortData converts Terraform's internal format to BIG-IP API format.
func unflattenFwRuleIpPortData(listVal []interface{}) *bigip.MgmtFwRuleIpPortData {
	if len(listVal) == 0 {
		return nil
	}

	if listVal[0] == nil {
		return nil
	}

	ipPortData := &bigip.MgmtFwRuleIpPortData{}
	data := listVal[0].(map[string]interface{})

	if addrLists, ok := data["address_lists"].([]interface{}); ok {
		ipPortData.AddressLists = make([]string, 0, len(addrLists))
		for _, item := range addrLists {
			ipPortData.AddressLists = append(ipPortData.AddressLists, item.(string))
		}
	}

	if addrs, ok := data["addresses"].(*schema.Set); ok {
		addrsList := addrs.List()
		ipPortData.Addresses = make([]bigip.MgmtFwRuleAddress, 0, len(addrsList))
		for _, item := range addrsList {
			addrMap := item.(map[string]interface{})
			ipPortData.Addresses = append(ipPortData.Addresses, bigip.MgmtFwRuleAddress{Name: addrMap["name"].(string)})
		}
	}

	if portLists, ok := data["port_lists"].([]interface{}); ok {
		ipPortData.PortLists = make([]string, 0, len(portLists))
		for _, item := range portLists {
			ipPortData.PortLists = append(ipPortData.PortLists, item.(string))
		}
	}

	if ports, ok := data["ports"].(*schema.Set); ok {
		portsList := ports.List()
		ipPortData.Ports = make([]bigip.MgmtFwRulePort, 0, len(portsList))
		for _, item := range portsList {
			portMap := item.(map[string]interface{})
			ipPortData.Ports = append(ipPortData.Ports, bigip.MgmtFwRulePort{Name: portMap["name"].(string)})
		}
	}

	return ipPortData
}

// buildManagementFwRuleConfig constructs a ManagementFirewallRule config from ResourceData
func buildManagementFwRuleConfig(d *schema.ResourceData) *bigip.MgmtFirewallRule {
	config := &bigip.MgmtFirewallRule{
		Name:   d.Get("name").(string),
		Action: d.Get("action").(string),
	}

	if val, ok := d.GetOk("log"); ok {
		config.Log = val.(string)
	}
	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}
	if val, ok := d.GetOk("status"); ok {
		config.Status = val.(string)
	}
	if val, ok := d.GetOk("ip_protocol"); ok {
		config.IpProtocol = val.(string)
	}
	if val, ok := d.GetOk("place_after"); ok {
		config.PlaceAfter = val.(string)
	}
	if val, ok := d.GetOk("place_before"); ok {
		config.PlaceBefore = val.(string)
	}
	if val, ok := d.GetOk("uuid"); ok {
		uuidVal := val.(string)
		if uuidVal == "auto-generate" {
			if d.Id() != "" {
				old, _ := d.GetChange("uuid")
				if oldUUID, ok := old.(string); ok && oldUUID != "" && oldUUID != "auto-generate" {
					config.UUID = oldUUID
				} else {
					config.UUID = "auto-generate"
				}
			} else {
				config.UUID = "auto-generate"
			}
		} else {
			config.UUID = uuidVal
		}
	}
	if val, ok := d.GetOk("schedule"); ok {
		config.Schedule = val.(string)
	}

	if val, ok := d.GetOk("source"); ok {
		sourceData := unflattenFwRuleIpPortData(val.([]interface{}))
		if sourceData != nil {
			config.Source = *sourceData
		} else {
			config.Source = bigip.MgmtFwRuleIpPortData{}
		}
	}

	if val, ok := d.GetOk("destination"); ok {
		destinationData := unflattenFwRuleIpPortData(val.([]interface{}))
		if destinationData != nil {
			config.Destination = *destinationData
		} else {
			config.Destination = bigip.MgmtFwRuleIpPortData{}
		}
	}

	if val, ok := d.GetOk("icmp"); ok {
		icmpList := val.(*schema.Set).List()
		config.ICMPs = make([]bigip.MgmtFwRuleICMP, 0, len(icmpList))
		for _, item := range icmpList {
			config.ICMPs = append(config.ICMPs, bigip.MgmtFwRuleICMP{
				Name: item.(map[string]interface{})["name"].(string),
			})
		}
	}

	return config
}

func resourceBigipSysMgmtFwRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating Management Firewall Rule: %s", name)
	config := buildManagementFwRuleConfig(d)

	if err := client.CreateManagementFwRule(config); err != nil {
		log.Printf("[ERROR] Unable to Create Management Fw Rule (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysMgmtFwRuleRead(ctx, d, meta)
}

func resourceBigipSysMgmtFwRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating Management Firewall Rule: %s", name)
	config := buildManagementFwRuleConfig(d)
	config.Name = name

	if err := client.ModifyManagementFwRule(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify Management Fw Rule (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysMgmtFwRuleRead(ctx, d, meta)
}

// flattenFwRuleIpPortData converts BIG-IP API format to Terraform's internal format
func flattenFwRuleIpPortData(data *bigip.MgmtFwRuleIpPortData) map[string]interface{} {
	out_data := make(map[string]interface{})
	if data == nil {
		return out_data
	}

	out_data["address_lists"] = data.AddressLists

	addresses := make([]interface{}, 0, len(data.Addresses))
	for _, item := range data.Addresses {
		addresses = append(addresses, map[string]interface{}{"name": item.Name})
	}
	out_data["addresses"] = addresses

	out_data["port_lists"] = data.PortLists

	ports := make([]interface{}, 0, len(data.Ports))
	for _, item := range data.Ports {
		ports = append(ports, map[string]interface{}{"name": item.Name})
	}
	out_data["ports"] = ports

	return out_data
}

// setAttrFwRuleWithDiag is a helper that returns diagnostics for error handling
func setAttrFwRuleWithDiag(d *schema.ResourceData, attrName string, val *string) diag.Diagnostics {
	if val != nil && *val != "" {
		if err := d.Set(attrName, *val); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceBigipSysMgmtFwRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading Mgmt Fw Rule config: %s", name)

	obj, err := client.GetManagementFwRule(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Management Fw Rule (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Management Fw Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", obj.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("action", obj.Action); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	diags = append(diags, setAttrFwRuleWithDiag(d, "log", &obj.Log)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "status", &obj.Status)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "ip_protocol", &obj.IpProtocol)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "place_after", &obj.PlaceAfter)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "place_before", &obj.PlaceBefore)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "description", &obj.Description)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "uuid", &obj.UUID)...)
	diags = append(diags, setAttrFwRuleWithDiag(d, "schedule", &obj.Schedule)...)

	if err := d.Set("source", []interface{}{flattenFwRuleIpPortData(&obj.Source)}); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("destination", []interface{}{flattenFwRuleIpPortData(&obj.Destination)}); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	icmps := make([]interface{}, 0, len(obj.ICMPs))
	for _, item := range obj.ICMPs {
		icmps = append(icmps, map[string]interface{}{
			"name": item.Name,
		})
	}
	if err := d.Set("icmp", icmps); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysMgmtFwRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Deleting Management Fw Rule " + name)

	err := client.DeleteManagementFwRule(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Management Fw Rule  (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
