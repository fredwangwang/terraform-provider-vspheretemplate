package vsphere_template

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"path/filepath"
	"time"
)

// defaultAPITimeout is a default timeout value that is passed to functions
// requiring contexts, and other various waiters.
const defaultAPITimeout = time.Minute * 60

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
			"allow_unverified_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, VMware vSphere client will permit unverifiable SSL certificates.",
			},
			"vcenter_server": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_VCENTER", nil),
				Deprecated:  "This field has been renamed to vsphere_server.",
			},
			"client_debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG", false),
				Description: "govmomi debug",
			},
			"client_debug_path_run": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG_PATH_RUN", ""),
				Description: "govmomi debug path for a single run",
			},
			"client_debug_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG_PATH", ""),
				Description: "govmomi debug path for debug",
			},
			"persist_session": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PERSIST_SESSION", false),
				Description: "Persist vSphere client sessions to disk",
			},
			"vim_session_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_VIM_SESSION_PATH", filepath.Join(os.Getenv("HOME"), ".govmomi", "sessions")),
				Description: "The directory to save vSphere SOAP API sessions to",
			},
			"rest_session_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_REST_SESSION_PATH", filepath.Join(os.Getenv("HOME"), ".govmomi", "rest_sessions")),
				Description: "The directory to save vSphere REST API sessions to",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vspheretemplate_ova_template": resourceVspheretemplateOvaTemplate(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	c, err := NewConfig(d)
	if err != nil {
		return nil, fmt.Errorf("failed to load the config: %s", err)
	}

	return c.Client()
}
