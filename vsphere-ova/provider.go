package vsphere_ova

import (
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_USER", nil),
				Description: "The user name for vSphere API operations.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PASSWORD", nil),
				Description: "The user password for vSphere API operations.",
			},
			"vsphere_server": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_SERVER", nil),
				Description: "The vSphere Server name for vSphere API operations.",
			},
			"vcenter_server": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_VCENTER", nil),
				Deprecated:  "This field has been renamed to vsphere_server.",
			},
			"allow_unverified_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, vSphere client will permit unverifiable SSL certificates.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vsphereova_ova_template": resourceVsphereovaOvaTemplate(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	server := d.Get("vsphere_server").(string)

	if server == "" {
		server = d.Get("vcenter_server").(string)
	}

	if server == "" {
		return nil, fmt.Errorf("one of vsphere_server or [deprecated] vcenter_server must be provided")
	}

	runner := &govcRunner{
		User:          d.Get("user").(string),
		Password:      d.Get("password").(string),
		InsecureFlag:  d.Get("allow_unverified_ssl").(bool),
		VSphereServer: server,
	}

	return runner, nil
}
