package virtualmachine

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"log"
)

// FromUUID locates a virtualMachine by its UUID.
func FromUUID(client *govmomi.Client, uuid string) (*object.VirtualMachine, error) {
	log.Printf("[DEBUG] Locating virtual machine with UUID %q", uuid)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var result object.Reference
	var err error

	result, err = object.NewSearchIndex(client.Client).FindByUuid(
		ctx, nil, uuid, true, nil)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	// We need to filter our object through finder to ensure that the
	// InventoryPath field is populated, or else functions that depend on this
	// being present will fail.
	finder := find.NewFinder(client.Client, false)

	vm, err := finder.ObjectReference(ctx, result.Reference())
	if err != nil {
		return nil, err
	}

	// Should be safe to return here. If our reference returned here and is not a
	// VM, then we have bigger problems and to be honest we should be panicking
	// anyway.
	log.Printf("[DEBUG] VM %q found for UUID %q", vm.(*object.VirtualMachine).InventoryPath, uuid)
	return vm.(*object.VirtualMachine), nil
}
