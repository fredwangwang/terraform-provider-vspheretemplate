package vsphere_ova

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"

	"bytes"
	"context"
	"errors"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova/archive"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova/datastore"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova/folder"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova/hostsystem"
	"github.com/fredwangwang/terraform-provider-vsphereova/vsphere-ova/resourcepool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"path"
)

func resourceVsphereovaOvaTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceVsphereovaOvaTemplateCreate,
		Read:   resourceVsphereovaOvaTemplateRead,
		Delete: resourceVsphereovaOvaTemplateDelete,

		Schema: map[string]*schema.Schema{
			"datastore_id": {
				Type:          schema.TypeString,
				Required:      true,
				ForceNew:      true,
				Description:   "The ID of the virtual machine's datastore. The virtual machine configuration is placed here, along with any virtual disks that are created without datastores.",
			},
			"folder": {
				Type:        schema.TypeString,
				Optional:    true,
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
			"options": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "If the ova file provided requires any special options to be set, set here.",
			},
			"resource_pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a resource pool to put the virtual machine in.",
			},
			"ova_file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "path to the ova file.",
			},
		},
	}
}

func resourceVsphereovaOvaTemplateCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	client := m.(*govmomi.Client)

	// retrieve iaas information
	ds, err := datastore.FromID(client, d.Get("datastore_id").(string))
	if err != nil {
		return fmt.Errorf("error locating datastore for VM: %s", err)
	}
	return fmt.Errorf("*************GOOD******************")
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

	ovaPath := d.Get("datastore_id").(string)
	archive := archive.NewTapeArchive(ovaPath, archive.Opener{Downloader: client})

	// load ova file
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
	vAppName := "Generic Virtual Appliance"
	if e.VirtualSystem != nil {
		vAppName = e.VirtualSystem.ID
		if e.VirtualSystem.Name != nil {
			vAppName = *e.VirtualSystem.Name
		}
	}

	//// Override vAppName from options if specified
	//if cmd.Options.Name != nil {
	//	vAppName = *cmd.Options.Name
	//}
	//
	//// Override vAppName from arguments if specified
	//if cmd.Name != "" {
	//	vAppName = cmd.Name
	//}

	cisp := types.OvfCreateImportSpecParams{
		//DiskProvisioning:   cmd.Options.DiskProvisioning,
		EntityName: vAppName,
		//IpAllocationPolicy: cmd.Options.IPAllocationPolicy,
		//IpProtocol:         cmd.Options.IPProtocol,
		//OvfManagerCommonParams: types.OvfManagerCommonParams{
		//	DeploymentOption: cmd.Options.Deployment,
		//	Locale:           "US"},
		//PropertyMapping: cmd.Map(cmd.Options.PropertyMapping),
		//NetworkMapping:  cmd.NetworkMap(e),
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

	//if cmd.Options.Annotation != "" {
	//	switch s := spec.ImportSpec.(type) {
	//	case *types.VirtualMachineImportSpec:
	//		s.ConfigSpec.Annotation = cmd.Options.Annotation
	//	case *types.VirtualAppImportSpec:
	//		s.VAppConfigSpec.Annotation = cmd.Options.Annotation
	//	}
	//}

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

	log.Printf("[INFO] Marking VM as template...\n")
	return vm.MarkAsTemplate(ctx)
}

//func importOva(archive importx.Archive, fpath string) (interface{}, error) {
//	//ctx := context.TODO()
//
//}

func upload(ctx context.Context, lease *nfc.Lease, item nfc.FileItem, archive archive.Archive) error {
	file := item.Path

	f, size, err := archive.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	tty, _ := flags.NewOutputFlag(ctx)
	logger := tty.ProgressLogger(fmt.Sprintf("Uploading %s... ", path.Base(file)))
	defer logger.Wait()

	opts := soap.Upload{
		ContentLength: size,
		Progress:      logger,
	}

	return lease.Upload(ctx, item, f, opts)
}

func resourceVsphereovaOvaTemplateRead(d *schema.ResourceData, m interface{}) error {
	return fmt.Errorf("ererer")
}

func resourceVsphereovaOvaTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
