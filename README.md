# Terraform-provider-vspheretemplate

Terraform vSphere provider does not provide the ability to upload the image file, 
so I created this toy provider to upload the ova image to vSphere.

## Provider Configuration:
There are three properties that are required: `user`, `password`, `vsphere_server`.
For more provider configuration, see the official
[terraform-provider-vsphere](https://www.terraform.io/docs/providers/vsphere/index.html#argument-reference)
documentation. This provider uses the same config as the official vsphere one.

## Resources:

* [vspheretemplate_ova_template](#vspheretemplate_ova_template)

## Resource Configuration:

### vspheretemplate_ova_template

* name - (Required) The name of the vm template.

* resource_pool_id - (Required) The managed object reference ID of the resource pool to put this vm template in.

* datastore_id - (Required) The managed object reference ID of the vm template's datastore

* datacenter_id - (Optional) The managed object reference ID of the vm template's datacenter. If there are more than one datacenter, this field is required.

* host_system_id - (Optional) An optional managed object reference ID of a host to put this vm template on. If a host_system_id is not supplied, vSphere will select a host in the resource pool to place the virtual machine, according to any defaults or DRS policies in place.

* [network_mapping](#network_mapping) - (Optional) Some image require network mapping. If you know your image needs a network_mapping or you get a `Host has no network defined` error when executing the script, you are required to provide this section

* [folder](#folder) - (Required) The path to the folder to put this virtual machine in, relative to the datacenter that the resource pool is in.

* ova_file_path - (Required) The path to the local ova file. (Theoretically you can provide a url to the ova file, not tested)

## Configuration Format:

### network_mapping:
sample:
```
// the format of the network_mapping is the same as the one in the `govc import.spec /path/to/image.ova`
network_mapping = {
	name = "vm network"
	network = "some-network"
}
```

### folder
Has to be in the format of `/datacenter/vm/actual_folder`

You can obtain the full path by browsing through the vsphere datacenter using `govc ls`

## Sample Script:
```hcl-terraform
provider "vspheretemplate" {
  //  uses thie same configuration as vsphere provider
  user = "${var.vcenter_user}"
  password = "${var.vcenter_password}"
  vsphere_server = "${var.vcenter_server}"

  allow_unverified_ssl = "${var.allow_unverified_ssl}"
}

data "vsphere_datacenter" "dc" {
  name = "${var.vcenter_dc}"
}

data "vsphere_datastore" "ds" {
  name = "${var.vcenter_ds}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_resource_pool" "pool" {
  name = "${var.vcenter_rp}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vspheretemplate_ova_template" "om_template" {
  name = "testing-template"
  datastore_id = "${data.vsphere_datastore.ds.id}"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  folder = "${var.vcenter_template_folder}"
  ova_file_path = "/path/to/image.ova"
}

variable "vcenter_user" {}

variable "vcenter_password" {}

variable "vcenter_server" {}

variable "allow_unverified_ssl" {}

variable "vcenter_dc" {}

variable "vcenter_ds" {}

variable "vcenter_rp" {}

variable "vcenter_template_folder" {}
```

## Issues
Of course when I created this provider, I want it to work alone with the official terraform-provider-vsphere
to provide a way to import ova and create vm in a single script. But the way terraform-provider-vsphere implements
does not allow that workflow. 

The `vsphere_virtual_machine` resource uses a custom diff function, which requires the `vm_template_uuid` be available
during the `refresh` stage (which happens before apply). But before the creation, the resource does not have an `id`,
which means 

```hcl-terraform
clone {
	template_uuid = "${vspheretemplate_ova_template.testing-template.id}"
}
```

does not work.

If you are interested, here is the trace:
```
resource_vsphere_virtual_machine.go:652
resource_vsphere_virtual_machine.go:708
virtual_machine_clone_subresource.go:60
virtual_machine_clone_subresource.go:65
```


## Credit
terraform-provider-vsphere: Borrowed `config.go` to talk to vsphere and generally how to code a provider

govmomi/govc: How to upload the ova file and how to use some of the govmomi apis.
