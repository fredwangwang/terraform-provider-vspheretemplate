package datacenter

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"log"
)

func FromIDOrDefault(client *govmomi.Client, id string) (*object.Datacenter, error) {
	log.Printf("[DEBUG] Locating datastore with ID %q", id)
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if id == "" {
		return finder.DefaultDatacenter(ctx)
	}

	ref := types.ManagedObjectReference{
		Type:  "Datacenter",
		Value: id,
	}

	dc, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Datastore with ID %q found", dc.Reference().Value)
	return dc.(*object.Datacenter), nil
}
