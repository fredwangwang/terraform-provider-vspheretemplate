package vsphere_template

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"

	"bytes"
	"context"
	"errors"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/archive"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/datacenter"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/datastore"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/folder"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/hostsystem"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/resourcepool"
	"github.com/fredwangwang/terraform-provider-vspheretemplate/vsphere-template/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVspheretemplateOvaTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceVspheretemplateOvaTemplateCreate,
		Read:   resourceVspheretemplateOvaTemplateRead,
		Delete: resourceVspheretemplateOvaTemplateDelete,

		Schema: map[string]*schema.Schema{
			"datacenter_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of the virtual machine's datacenter.",
			},
			"datastore_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the virtual machine's datastore. The virtual machine configuration is placed here, along with any virtual disks that are created without datastores.",
			},
			"folder": {
				Type:        schema.TypeString,
				// TODO: make this optional
				Required:    true,
				ForceNew:    true,
				Description: "The name of the folder to locate the virtual machine in.",
				//StateFunc:   folder.NormalizePath,
			},
			"host_system_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The ID of an optional host system to pin the virtual machine to.",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The display name of the template.",
				ForceNew:    true,
				Required:    true,
			},
			"network_mapping": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "If the ova file provided requires any network mapping to be set, set here.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"resource_pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a resource pool to put the virtual machine in.",
			},
			"ova_file_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "path to the ova file.",
			},
			"guest_id": {
				Type:        schema.TypeString,
				Description: "The guest ID of the virtual machine.",
				Computed:    true,
			},
		},
	}
}

func resourceVspheretemplateOvaTemplateCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	client := m.(*govmomi.Client)

	// retrieve iaas information
	ds, err := datastore.FromID(client, d.Get("datastore_id").(string))
	if err != nil {
		return fmt.Errorf("error locating datastore for VM: %s", err)
	}

	var hs *object.HostSystem
	if v, ok := d.GetOk("host_system_id"); ok {
		hsID := v.(string)
		var err error
		if hs, err = hostsystem.FromID(client, hsID); err != nil {
			return fmt.Errorf("error locating host system at ID %q: %s", hsID, err)
		}
	}

	poolID := d.Get("resource_pool_id").(string)
	pool, err := resourcepool.FromID(client, poolID)
	if err != nil {
		return fmt.Errorf("could not find resource pool ID %q: %s", poolID, err)
	}

	ovaPath := d.Get("ova_file_path").(string)
	archive := archive.NewTapeArchive(ovaPath, archive.Opener{Downloader: client})

	// load ovf file
	reader, _, err := archive.Open("*.ovf")
	if err != nil {
		return err
	}

	defer reader.Close()
	ovfContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read ovf: %s", err)
	}
	e, err := ovf.Unmarshal(bytes.NewReader(ovfContent))
	if err != nil {
		return fmt.Errorf("failed to parse ovf: %s", err)
	}

	// set appliance properties
	cisp, err := createImportSpecParams(d, e, client)
	if err != nil {
		return err
	}

	ovfManager := ovf.NewManager(client.Client)

	spec, err := ovfManager.CreateImportSpec(
		ctx, string(ovfContent),
		pool, ds, cisp)
	if err != nil {
		return err
	}
	if spec.Error != nil {
		return errors.New(spec.Error[0].LocalizedMessage)
	}
	if spec.Warning != nil {
		for _, w := range spec.Warning {
			log.Printf("[WARN] %s\n", w.LocalizedMessage)
		}
	}

	// TODO: add annotation @ importx/ovf.go:288

	folder, err := folder.FromName(client, d.Get("folder").(string))
	if err != nil {
		return err
	}

	lease, err := pool.ImportVApp(ctx, spec.ImportSpec, folder, hs)
	if err != nil {
		return err
	}

	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		return err
	}

	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	for _, i := range info.Items {
		err = upload(ctx, lease, i, archive)
		if err != nil {
			return err
		}
	}

	moref := info.Entity
	err = lease.Complete(ctx)
	if err != nil {
		return err
	}

	vm := object.NewVirtualMachine(client.Client, moref)
	d.SetId(vm.UUID(ctx))

	hs, err = vm.HostSystem(ctx)
	if err != nil {
		return err
	}
	d.Set("host_system_id", hs.Reference().Value)

	props, err := virtualmachine.Properties(vm)
	if err != nil {
		return fmt.Errorf("error fetching virtual machine properties: %s", err)
	}
	if props.Config == nil {
		return fmt.Errorf("no configuration returned for virtual machine %q", vm.InventoryPath)
	}

	d.Set("guest_id", props.Config.GuestId)

	log.Printf("[INFO] Marking VM as template...\n")
	return vm.MarkAsTemplate(ctx)
}

func createImportSpecParams(
	d *schema.ResourceData,
	envelope *ovf.Envelope,
	c *govmomi.Client) (types.OvfCreateImportSpecParams, error) {
	var err error

	vAppName := d.Get("name").(string)

	networkMapping := func(e *ovf.Envelope) (p []types.OvfNetworkMapping) {
		ctx := context.TODO()
		finder := find.NewFinder(c.Client, false)
		dc, err1 := datacenter.FromIDOrDefault(c, d.Get("datacenter_id").(string))
		if err1 != nil {
			err = err1
			return
		}
		finder.SetDatacenter(dc)

		networks := map[string]string{}

		if e.Network != nil {
			for _, net := range e.Network.Networks {
				networks[net.Name] = net.Name
			}
		}

		networkMappings := d.Get("network_mapping").([]interface{})

		for _, net := range networkMappings {
			netMap := net.(map[string]interface{})
			networks[netMap["name"].(string)] = netMap["network"].(string)
		}

		for src, dst := range networks {
			if net, err1 := finder.Network(ctx, dst); err1 == nil {
				p = append(p, types.OvfNetworkMapping{
					Name:    src,
					Network: net.Reference(),
				})
			} else {
				err = err1
				return
			}
		}
		return
	}

	cisp := types.OvfCreateImportSpecParams{
		EntityName:     vAppName,
		NetworkMapping: networkMapping(envelope),
	}

	return cisp, err
}

func upload(ctx context.Context, lease *nfc.Lease, item nfc.FileItem, archive archive.Archive) error {
	file := item.Path

	f, size, err := archive.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	opts := soap.Upload{
		ContentLength: size,
	}

	return lease.Upload(ctx, item, f, opts)
}

func resourceVspheretemplateOvaTemplateRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*govmomi.Client)

	vm, err := virtualmachine.FromUUID(client, d.Id())
	if err != nil {
		return err
	}

	if vm == nil {
		d.SetId("")
	}

	return nil
}

func resourceVspheretemplateOvaTemplateDelete(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	client := m.(*govmomi.Client)

	id := d.Id()

	vm, err := virtualmachine.FromUUID(client, id)
	if err != nil || vm == nil {
		return fmt.Errorf("cannot locate virtual machine with UUID %q", id)
	}

	task, err := vm.Destroy(ctx)
	if err != nil {
		return err
	}

	err = task.Wait(ctx)
	if err != nil {
		return err
	}

	d.SetId("")
	log.Printf("[DEBUG] %q: Delete complete", id)
	return nil
}
