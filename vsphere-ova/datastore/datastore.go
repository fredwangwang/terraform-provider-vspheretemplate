package datastore

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"log"
)

func FromID(client *govmomi.Client, id string) (*object.Datastore, error) {
	log.Printf("[DEBUG] Locating datastore with ID %q", id)
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "Datastore",
		Value: id,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ds, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Datastore with ID %q found", ds.Reference().Value)
	return ds.(*object.Datastore), nil
}
