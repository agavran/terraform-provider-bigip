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
	"strings"
	"time"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func httpdPatchWithRetry(client *bigip.BigIP, crudOpName string, config *bigip.HTTPDConfig) error {
	const maxRetries = 5
	const retryDelay = 15 * time.Second
	for i := 0; i < maxRetries; i++ {
		log.Printf("[DEBUG] HTTPD %s: attempt %d/%d", crudOpName, i+1, maxRetries)

		var err error
		switch crudOpName {
		case "create":
			err = client.CreateHTTPDConfig(config)
		case "modify":
			err = client.ModifyHTTPDConfig(config)
		case "delete":
			err = client.DeleteHTTPDConfig()
		default:
			return fmt.Errorf("HTTPD %s: unknown operation", crudOpName)
		}

		if err == nil {
			log.Printf("[DEBUG] HTTPD %s: attempt %d succeeded", crudOpName, i+1)
			return nil
		}
		log.Printf("[DEBUG] HTTPD %s: attempt %d error type=%T value=%v", crudOpName, i+1, err, err)
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "connection reset") {
			log.Printf("[INFO] HTTPD %s: httpd is restarting (%v), waiting %s before retry (%d/%d)", crudOpName, err, retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
			continue
		}
		return err
	}
	return fmt.Errorf("HTTPD %s: exceeded %d retries", crudOpName, maxRetries)
}

// defaultAllowAll returns the default value for the allow field
func defaultAllowAll() (interface{}, error) {
	return []interface{}{"All"}, nil
}

func resourceBigipSysHttpd() *schema.Resource {
	return &schema.Resource{
		Description: "Manages BIG-IP HTTP daemon configuration. " +
			"NOTE: Only one instance of this resource should exist per BIG-IP device. " +
			"IMPORTANT: F5 Networks recommends that users of the Configuration utility exit the utility before changes are made to the system using the httpd component." +
			"This is because making changes to the system using this component causes a restart of the httpd daemon." +
			"Additionally, restarting the httpd daemon creates the necessity for a restart of the Configuration utility.",
		CreateContext: resourceBigipSysHttpdCreate,
		UpdateContext: resourceBigipSysHttpdUpdate,
		ReadContext:   resourceBigipSysHttpdRead,
		DeleteContext: resourceBigipSysHttpdDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"allow": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of IP addresses, partial IPs, IP ranges, hostnames, domain names, or network/netmask pairs allowed to access the web interface (default: ALL)",
			},
			"auth_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "BIG-IP",
				Description: "Authentication realm name (default: BIG-IP)",
			},
			"auth_pam_dashboard_timeout": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Description:  "Enable dashboard timeout for PAM authentication (default: off)",
			},
			"auth_pam_idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1200,
				ValidateFunc: validation.IntBetween(120, 2147483647),
				Description:  "PAM authentication idle timeout in seconds (120-2147483647, default: 1200)",
			},
			"auth_pam_validate_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "on",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Description:  "Validate client IP for PAM authentication (default: on)",
			},
			"fastcgi_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "FastCGI timeout in seconds (default: 300)",
			},
			"fips_cipher_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "FIPS cipher version - read-only, set by system based on FIPS mode",
			},
			"hostname_lookup": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Description:  "Enable hostname lookups in logs (default: off)",
			},
			"include": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "Include custom HTTPD configuration file",
			},
			"log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "warn",
				ValidateFunc: validation.StringInSlice([]string{"alert", "crit", "debug", "emerg", "error", "info", "notice", "warn"}, false),
				Description:  "HTTPD log level (default: warn)",
			},
			"max_clients": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10,
				ValidateFunc: validation.IntBetween(10, 256),
				Description:  "Maximum number of concurrent clients (10-256, default: 10)",
			},
			"redirect_http_to_https": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Redirect HTTP requests to HTTPS (default: disabled)",
			},
			"request_body_max_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum timeout for receiving request body (default: 0)",
			},
			"request_body_min_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      500,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Minimum rate for receiving request body in bytes/sec (default: 500)",
			},
			"request_body_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Timeout for receiving request body in seconds (default: 60)",
			},
			"request_header_max_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      40,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum timeout for receiving request headers (default: 40)",
			},
			"request_header_min_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      500,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Minimum rate for receiving request headers in bytes/sec (default: 500)",
			},
			"request_header_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Timeout for receiving request headers in seconds (default: 20)",
			},
			"ssl_ca_cert_file": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "SSL CA certificate file path",
			},
			"ssl_certchainfile": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "SSL certificate chain file path",
			},
			"ssl_certfile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/etc/httpd/conf/ssl.crt/server.crt",
				Description: "SSL certificate file path",
			},
			"ssl_certkeyfile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/etc/httpd/conf/ssl.key/server.key",
				Description: "SSL certificate key file path",
			},
			"ssl_ciphersuite": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-SHA:ECDHE-ECDSA-AES256-SHA:ECDHE-ECDSA-AES128-SHA256:ECDHE-ECDSA-AES256-SHA384:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA:AES256-SHA:AES128-SHA256:AES256-SHA256",
				Description: "SSL cipher suite string",
			},
			"ssl_include": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressNoneDiff,
				Description:      "Include custom SSL configuration file",
			},
			"ssl_ocsp_default_responder": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "http://127.0.0.1",
				Description: "Default OCSP responder URL",
			},
			"ssl_ocsp_enable": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Description:  "Enable OCSP validation (default: off)",
			},
			"ssl_ocsp_override_responder": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Description:  "Override OCSP responder from certificate (default: off)",
			},
			"ssl_ocsp_responder_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "OCSP responder timeout in seconds (default: 300)",
			},
			"ssl_ocsp_response_max_age": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Maximum age of OCSP response in seconds (-1 for no limit, default: -1)",
			},
			"ssl_ocsp_response_time_skew": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "OCSP response time skew tolerance in seconds (default: 300)",
			},
			"ssl_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      443,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "HTTPS listening port (default: 443)",
			},
			"ssl_protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "all -SSLv2 -SSLv3 -TLSv1",
				Description: "SSL/TLS protocol versions to enable (default: all -SSLv2 -SSLv3 -TLSv1)",
			},
			"ssl_verify_client": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "no",
				ValidateFunc: validation.StringInSlice([]string{"no", "optional", "require", "optional_no_ca"}, false),
				Description:  "Client certificate verification mode (default: no)",
			},
			"ssl_verify_depth": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10,
				ValidateFunc: validation.IntBetween(10, 2147483647),
				Description:  "Maximum depth of CA certificates chain verification (10-2147483647, default: 10)",
			},
		},
	}
}

// buildHttpdConfig constructs an HTTPDConfig from ResourceData
func buildHttpdConfig(d *schema.ResourceData) *bigip.HTTPDConfig {
	config := &bigip.HTTPDConfig{}

	if val, ok := d.GetOk("allow"); ok {
		allowList := val.([]interface{})
		allow := make([]string, len(allowList))
		for i, v := range allowList {
			allow[i] = v.(string)
		}
		config.Allow = allow
	}

	setStringIfOk(d, "auth_name", &config.AuthName)
	setStringIfOk(d, "auth_pam_dashboard_timeout", &config.AuthPamDashboardTimeout)
	setStringIfOk(d, "auth_pam_validate_ip", &config.AuthPamValidateIp)
	setStringIfOk(d, "hostname_lookup", &config.HostnameLookup)
	setStringIfOk(d, "include", &config.Include)
	setStringIfOk(d, "log_level", &config.LogLevel)
	setStringIfOk(d, "redirect_http_to_https", &config.RedirectHttpToHttps)
	setStringIfOk(d, "ssl_ca_cert_file", &config.SslCaCertFile)
	setStringIfOk(d, "ssl_certchainfile", &config.SslCertchainfile)
	setStringIfOk(d, "ssl_certfile", &config.SslCertfile)
	setStringIfOk(d, "ssl_certkeyfile", &config.SslCertkeyfile)
	setStringIfOk(d, "ssl_ciphersuite", &config.SslCiphersuite)
	setStringIfOk(d, "ssl_include", &config.SslInclude)
	setStringIfOk(d, "ssl_ocsp_default_responder", &config.SslOcspDefaultResponder)
	setStringIfOk(d, "ssl_ocsp_enable", &config.SslOcspEnable)
	setStringIfOk(d, "ssl_ocsp_override_responder", &config.SslOcspOverrideResponder)
	setStringIfOk(d, "ssl_protocol", &config.SslProtocol)
	setStringIfOk(d, "ssl_verify_client", &config.SslVerifyClient)

	if val, ok := d.GetOk("auth_pam_idle_timeout"); ok {
		config.AuthPamIdleTimeout = val.(int)
	}
	if val, ok := d.GetOk("fastcgi_timeout"); ok {
		config.FastcgiTimeout = val.(int)
	}
	if val, ok := d.GetOk("max_clients"); ok {
		config.MaxClients = val.(int)
	}
	if val, ok := d.GetOk("request_body_max_timeout"); ok {
		config.RequestBodyMaxTimeout = val.(int)
	}
	if val, ok := d.GetOk("request_body_min_rate"); ok {
		config.RequestBodyMinRate = val.(int)
	}
	if val, ok := d.GetOk("request_body_timeout"); ok {
		config.RequestBodyTimeout = val.(int)
	}
	if val, ok := d.GetOk("request_header_max_timeout"); ok {
		config.RequestHeaderMaxTimeout = val.(int)
	}
	if val, ok := d.GetOk("request_header_min_rate"); ok {
		config.RequestHeaderMinRate = val.(int)
	}
	if val, ok := d.GetOk("request_header_timeout"); ok {
		config.RequestHeaderTimeout = val.(int)
	}
	if val, ok := d.GetOk("ssl_ocsp_responder_timeout"); ok {
		config.SslOcspResponderTimeout = val.(int)
	}
	if val, ok := d.GetOk("ssl_ocsp_response_max_age"); ok {
		config.SslOcspResponseMaxAge = val.(int)
	}
	if val, ok := d.GetOk("ssl_ocsp_response_time_skew"); ok {
		config.SslOcspResponseTimeSkew = val.(int)
	}
	if val, ok := d.GetOk("ssl_port"); ok {
		config.SslPort = val.(int)
	}
	if val, ok := d.GetOk("ssl_verify_depth"); ok {
		config.SslVerifyDepth = val.(int)
	}

	return config
}

func resourceBigipSysHttpdCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating HTTPD configuration")
	config := buildHttpdConfig(d)

	err := httpdPatchWithRetry(client, "create", config)
	if err != nil {
		log.Printf("[ERROR] Unable to Create HTTPD configuration: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("httpd_config")
	return resourceBigipSysHttpdRead(ctx, d, meta)
}

func resourceBigipSysHttpdUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating HTTPD configuration")
	config := buildHttpdConfig(d)

	if d.HasChange("include") {
		if _, ok := d.GetOk("include"); !ok {
			log.Printf("[INFO] Include removed from config, clearing it")
			config.Include = "none"
		}
	}
	if d.HasChange("ssl_include") {
		if _, ok := d.GetOk("ssl_include"); !ok {
			log.Printf("[INFO] SSL include removed from config, clearing it")
			config.SslInclude = "none"
		}
	}
	if d.HasChange("ssl_ca_cert_file") {
		if _, ok := d.GetOk("ssl_ca_cert_file"); !ok {
			log.Printf("[INFO] SSL CA cert file removed from config, clearing it")
			config.SslCaCertFile = "none"
		}
	}
	if d.HasChange("ssl_certchainfile") {
		if _, ok := d.GetOk("ssl_certchainfile"); !ok {
			log.Printf("[INFO] SSL cert chain file removed from config, clearing it")
			config.SslCertchainfile = "none"
		}
	}

	err := httpdPatchWithRetry(client, "modify", config)
	if err != nil {
		log.Printf("[ERROR] Unable to Modify HTTPD configuration: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysHttpdRead(ctx, d, meta)
}

func resourceBigipSysHttpdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading HTTPD configuration")

	obj, err := client.GetHTTPDConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve HTTPD configuration: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] HTTPD configuration not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	setResourceData(d, "allow", obj.Allow, &diags)
	setResourceData(d, "auth_name", obj.AuthName, &diags)
	setResourceData(d, "auth_pam_dashboard_timeout", obj.AuthPamDashboardTimeout, &diags)
	setResourceData(d, "auth_pam_idle_timeout", obj.AuthPamIdleTimeout, &diags)
	setResourceData(d, "auth_pam_validate_ip", obj.AuthPamValidateIp, &diags)
	setResourceData(d, "fastcgi_timeout", obj.FastcgiTimeout, &diags)
	setResourceData(d, "fips_cipher_version", obj.FipsCipherVersion, &diags)
	setResourceData(d, "hostname_lookup", obj.HostnameLookup, &diags)
	setResourceData(d, "include", obj.Include, &diags)
	setResourceData(d, "log_level", obj.LogLevel, &diags)
	setResourceData(d, "max_clients", obj.MaxClients, &diags)
	setResourceData(d, "redirect_http_to_https", obj.RedirectHttpToHttps, &diags)
	setResourceData(d, "request_body_max_timeout", obj.RequestBodyMaxTimeout, &diags)
	setResourceData(d, "request_body_min_rate", obj.RequestBodyMinRate, &diags)
	setResourceData(d, "request_body_timeout", obj.RequestBodyTimeout, &diags)
	setResourceData(d, "request_header_max_timeout", obj.RequestHeaderMaxTimeout, &diags)
	setResourceData(d, "request_header_min_rate", obj.RequestHeaderMinRate, &diags)
	setResourceData(d, "request_header_timeout", obj.RequestHeaderTimeout, &diags)
	setResourceData(d, "ssl_ca_cert_file", obj.SslCaCertFile, &diags)
	setResourceData(d, "ssl_certchainfile", obj.SslCertchainfile, &diags)
	setResourceData(d, "ssl_certfile", obj.SslCertfile, &diags)
	setResourceData(d, "ssl_certkeyfile", obj.SslCertkeyfile, &diags)
	setResourceData(d, "ssl_ciphersuite", obj.SslCiphersuite, &diags)
	setResourceData(d, "ssl_include", obj.SslInclude, &diags)
	setResourceData(d, "ssl_ocsp_default_responder", obj.SslOcspDefaultResponder, &diags)
	setResourceData(d, "ssl_ocsp_enable", obj.SslOcspEnable, &diags)
	setResourceData(d, "ssl_ocsp_override_responder", obj.SslOcspOverrideResponder, &diags)
	setResourceData(d, "ssl_ocsp_responder_timeout", obj.SslOcspResponderTimeout, &diags)
	setResourceData(d, "ssl_ocsp_response_max_age", obj.SslOcspResponseMaxAge, &diags)
	setResourceData(d, "ssl_ocsp_response_time_skew", obj.SslOcspResponseTimeSkew, &diags)
	setResourceData(d, "ssl_port", obj.SslPort, &diags)
	setResourceData(d, "ssl_protocol", obj.SslProtocol, &diags)
	setResourceData(d, "ssl_verify_client", obj.SslVerifyClient, &diags)
	setResourceData(d, "ssl_verify_depth", obj.SslVerifyDepth, &diags)

	return diags
}

func resourceBigipSysHttpdDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting HTTPD configuration to defaults")

	err := httpdPatchWithRetry(client, "delete", nil)
	if err != nil {
		log.Printf("[ERROR] Unable to Reset HTTPD configuration (%v)", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
