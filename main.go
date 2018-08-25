package main

import (
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vsphere_ova.Provider,
	})
}
