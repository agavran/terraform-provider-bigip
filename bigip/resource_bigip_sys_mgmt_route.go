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

// validateNetworkCIDR validates network CIDR notation or "default"
func validateNetworkCIDR(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "default" {
		return // "default" is valid for default route
	}
	_, _, err := net.ParseCIDR(v)
	if err != nil {
		errs = append(errs, fmt.Errorf("%q must be a valid CIDR notation or 'default', got: %s", key, v))
	}
	return
}

func resourceBigipSysMgmtRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysMgmtRouteCreate,
		UpdateContext: resourceBigipSysMgmtRouteUpdate,
		ReadContext:   resourceBigipSysMgmtRouteRead,
		DeleteContext: resourceBigipSysMgmtRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5Name,
				Description:  "Name of the management route",
			},
			"network": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateNetworkCIDR,
				Description:  "Destination network in CIDR notation (e.g., '10.0.0.0/24') or 'default' for default route",
			},
			"gateway": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Gateway IP address for the route",
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1500,
				ValidateFunc: validation.IntBetween(0, 9198),
				Description:  "MTU for the management route traffic (default: 1500, 0 means use system default)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the management route",
			},
		},
	}
}

// buildManagementRouteConfig constructs a ManagementRoute config from ResourceData
func buildManagementRouteConfig(d *schema.ResourceData) *bigip.ManagementRoute {
	config := &bigip.ManagementRoute{
		Name:    d.Get("name").(string),
		Network: d.Get("network").(string),
		Gateway: d.Get("gateway").(string),
	}

	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}
	if val, ok := d.GetOk("mtu"); ok {
		config.MTU = val.(int)
	}

	return config
}

func resourceBigipSysMgmtRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating Management Route: %s", name)
	config := buildManagementRouteConfig(d)

	if err := client.CreateManagementRoute(config); err != nil {
		log.Printf("[ERROR] Unable to Create Management Route (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysMgmtRouteRead(ctx, d, meta)
}

func resourceBigipSysMgmtRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating Management Route: %s", name)
	config := buildManagementRouteConfig(d)
	config.Name = name // Ensure we use the ID for updates

	if err := client.ModifyManagementRoute(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify Management Route (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysMgmtRouteRead(ctx, d, meta)
}

func resourceBigipSysMgmtRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading Mgmt Route config: %s", name)

	obj, err := client.GetManagementRoute(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Management Route (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Management Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", obj.FullPath); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("network", obj.Network); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("gateway", obj.Gateway); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("mtu", obj.MTU); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", obj.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysMgmtRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting Management Route: %s", name)

	err := client.DeleteManagementRoute(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Management Route  (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
