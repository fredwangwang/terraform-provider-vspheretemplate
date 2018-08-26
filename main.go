package main

import (
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vsphere_template.Provider,
	})
}
