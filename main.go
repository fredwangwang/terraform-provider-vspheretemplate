package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vsphere_ova.Provider,
	})
}
