/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"context"
	"fmt"
	"log"
	"time"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBigipSysProvision() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysProvisionCreate,
		UpdateContext: resourceBigipSysProvisionUpdate,
		ReadContext:   resourceBigipSysProvisionRead,
		DeleteContext: resourceBigipSysProvisionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of module to provision in BIG-IP.",
				ValidateFunc: validation.StringInSlice([]string{"afm", "am", "apm", "asm", "avr", "cgnat", "dos", "fps", "gtm", "ilx", "lc", "ltm", "pem", "sslo", "swg", "urldb"}, false),
			},
			"full_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				//ValidateFunc: validation.StringInSlice([]string{"afm", "am", "apm","asm","avr","dos","fps","gtm","ilx","lc","ltm","pem", "sslo" ,"swg","urldb"}, false),
			},
			"cpu_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Use this option only when the level option is set to custom.F5 Networks recommends that you do not modify this option. The default value is none",
			},
			"disk_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Use this option only when the level option is set to custom.F5 Networks recommends that you do not modify this option. The default value is none",
			},
			"level": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Sets the provisioning level for the requested modules. Changing the level for one module may require modifying the level of another module. For example, changing one module to dedicated requires setting all others to none. Setting the level of a module to none means the module is not activated.",
				Default:      "nominal",
				ValidateFunc: validation.StringInSlice([]string{"nominal", "none", "minimum", "dedicated"}, false),
			},
			"memory_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Use this option only when the level option is set to custom.F5 Networks recommends that you do not modify this option. The default value is none",
			},
		},
	}
}

func resourceBigipSysProvisionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Provisioning for %v module", name)

	pss := &bigip.Provision{
		Name: name,
	}
	config := getsysProvisionConfig(d, pss)

	err := client.ProvisionModule(config)
	if err != nil {
		log.Printf("[ERROR] Unable to Create Provision  (%s) ", err)
		return diag.FromErr(err)
	}
	d.SetId(name)

	// Provisioning causes service restarts. Wait and retry the read operation
	// to handle transient 502/503 errors during service restart.
	return resourceBigipSysProvisionReadWithRetry(ctx, d, meta)
}

func resourceBigipSysProvisionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Updating Provisioning for :%v module", name)

	pss := &bigip.Provision{
		Name: name,
	}
	config := getsysProvisionConfig(d, pss)

	err := client.ProvisionModule(config)
	if err != nil {
		log.Printf("[ERROR] Unable to Update Provision (%v) ", err)
		return diag.FromErr(err)
	}

	// Provisioning causes service restarts. Wait and retry the read operation
	// to handle transient 502/503 errors during service restart.
	return resourceBigipSysProvisionReadWithRetry(ctx, d, meta)
}

// resourceBigipSysProvisionReadWithRetry handles the read operation with retry logic
// to accommodate BIG-IP service restarts that occur after provisioning changes.
// It will poll until the API is available (up to 2 minutes), then perform a final read.
func resourceBigipSysProvisionReadWithRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var lastErr error
	maxRetries := 40
	retryDelay := 3 * time.Second
	initialDelay := 15 * time.Second
	levelConvergeRetries := 5

	// Capture the desired level before the retry loop, since successful reads
	// will overwrite ResourceData with the current (possibly stale) API value.
	desiredLevel := d.Get("level").(string)

	log.Printf("[INFO] Waiting for BIG-IP services to stabilize after provisioning...")

	log.Printf("[INFO] Initial delay of %v to allow service restart to begin...", initialDelay)
	time.Sleep(initialDelay)

	// Poll until the API responds successfully or timeout
	apiAvailable := false
	for i := 0; i < maxRetries; i++ {
		diags := resourceBigipSysProvisionRead(ctx, d, meta)
		if !diags.HasError() {
			if i > 0 {
				log.Printf("[INFO] BIG-IP API became available after %d attempts", i+1)
			}
			apiAvailable = true
			break
		}

		// Continue retrying on ANY error until timeout
		errMsg := diags[0].Summary
		lastErr = fmt.Errorf("%s", errMsg)
		log.Printf("[WARN] Provision read attempt %d/%d failed: %s. Retrying...", i+1, maxRetries, errMsg)
		time.Sleep(retryDelay)
	}

	if !apiAvailable {
		return diag.FromErr(fmt.Errorf("provision read failed after %d retries. Last error: %v", maxRetries, lastErr))
	}

	// API is available - verify that the provisioned level matches the desired level.
	// BIG-IP may return 200 OK incorrect or unexpcted data, trying few times.
	for i := 0; i < levelConvergeRetries; i++ {
		currentLevel := d.Get("level").(string)
		if currentLevel == desiredLevel {
			log.Printf("[INFO] Provision level converged to desired value: %s", desiredLevel)
			return nil
		}

		log.Printf("[WARN] Provision level mismatch (attempt %d/%d): got %q, want %q. Retrying...",
			i+1, levelConvergeRetries, currentLevel, desiredLevel)
		time.Sleep(retryDelay)

		diags := resourceBigipSysProvisionRead(ctx, d, meta)
		if diags.HasError() {
			return diags
		}
	}

	// Final check after all retries
	currentLevel := d.Get("level").(string)
	if currentLevel != desiredLevel {
		return diag.FromErr(fmt.Errorf("provision level did not converge: got %q, want %q after %d attempts",
			currentLevel, desiredLevel, levelConvergeRetries))
	}

	return nil
}

func resourceBigipSysProvisionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Println("[INFO] Reading Provisions " + name)
	p, err := client.Provisions(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Provision (%s) (%v) ", name, err)
		return diag.FromErr(err)
	}
	if p == nil {
		log.Printf("[WARN] Provision (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	_ = d.Set("full_path", p.FullPath)
	_ = d.Set("cpu_ratio", p.CpuRatio)
	_ = d.Set("disk_ratio", p.DiskRatio)
	_ = d.Set("level", p.Level)
	_ = d.Set("memory_ratio", p.MemoryRatio)

	return nil
}

func resourceBigipSysProvisionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// API is not supported for Deleting
	return nil
}

func getsysProvisionConfig(d *schema.ResourceData, config *bigip.Provision) *bigip.Provision {
	config.FullPath = d.Get("full_path").(string)
	config.CpuRatio = d.Get("cpu_ratio").(int)
	config.DiskRatio = d.Get("disk_ratio").(int)
	config.Level = d.Get("level").(string)
	config.MemoryRatio = d.Get("memory_ratio").(int)
	return config
}
