package bicc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tra-jkv/terraform-provider-bicc/bicc/client"
)

// Provider returns the BICC Terraform provider schema
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BICC_HOST", nil),
				Description: "The Oracle Fusion Applications hostname (e.g., servername.fa.us2.oraclecloud.com)",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BICC_USERNAME", nil),
				Description: "Username for BICC authentication",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("BICC_PASSWORD", nil),
				Description: "Password for BICC authentication",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     443,
				Description: "Port for BICC API (default: 443)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bicc_job":          resourceBICCJob(),
			"bicc_job_backfill": resourceBICCJobBackfill(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	port := d.Get("port").(int)

	config := &client.Config{
		Host:     host,
		Username: username,
		Password: password,
		Port:     port,
	}

	biccClient := client.NewClient(config)

	return biccClient, diags
}
