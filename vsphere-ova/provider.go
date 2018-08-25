package vsphere_ova

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-vsphere/vsphere"
	"path/filepath"
	"os"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_USER", nil),
				Description: "The user name for vSphere API operations.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PASSWORD", nil),
				Description: "The user password for vSphere API operations.",
			},

			"vsphere_server": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_SERVER", nil),
				Description: "The vSphere Server name for vSphere API operations.",
			},
			"allow_unverified_ssl": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, VMware vSphere client will permit unverifiable SSL certificates.",
			},
			"vcenter_server": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_VCENTER", nil),
				Deprecated:  "This field has been renamed to vsphere_server.",
			},
			"client_debug": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG", false),
				Description: "govmomi debug",
			},
			"client_debug_path_run": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG_PATH_RUN", ""),
				Description: "govmomi debug path for a single run",
			},
			"client_debug_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_CLIENT_DEBUG_PATH", ""),
				Description: "govmomi debug path for debug",
			},
			"persist_session": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PERSIST_SESSION", false),
				Description: "Persist vSphere client sessions to disk",
			},
			"vim_session_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_VIM_SESSION_PATH", filepath.Join(os.Getenv("HOME"), ".govmomi", "sessions")),
				Description: "The directory to save vSphere SOAP API sessions to",
			},
			"rest_session_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_REST_SESSION_PATH", filepath.Join(os.Getenv("HOME"), ".govmomi", "rest_sessions")),
				Description: "The directory to save vSphere REST API sessions to",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vsphereova_ova_template": resourceVsphereovaOvaTemplate(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	c, err :=  vsphere.NewConfig(d)
	client, err := c.Client()
	return client, err
	//c, err := vsphere.NewConfig(d)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to load the config: %s", err)
	//}
	//
	//client, err := c.Client()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to retrive a govmomi client: %s", err)
	//}
	//
	//log.Println(err, client)
	//v := reflect.ValueOf(*client)
	//for i := 0; i < v.NumField(); i++ {
	//	if v.Field(i).Type() == reflect.TypeOf(&govmomi.Client{}) {
	//		return v.Field(i).Elem().Interface().(*govmomi.Client), nil
	//	}
	//}
	////
	//return nil, fmt.Errorf("failed to retrive a govmomi client")
}
