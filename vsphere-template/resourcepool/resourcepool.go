package resourcepool

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"log"
)

func FromID(client *govmomi.Client, id string) (*object.ResourcePool, error) {
	log.Printf("[DEBUG] Locating resource pool with ID %s", id)
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "ResourcePool",
		Value: id,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	obj, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Resource pool found: %s", obj.Reference().Value)
	return obj.(*object.ResourcePool), nil
}
