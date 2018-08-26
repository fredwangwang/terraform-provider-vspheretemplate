package hostsystem

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"log"
)

func FromID(client *govmomi.Client, id string) (*object.HostSystem, error) {
	log.Printf("[DEBUG] Locating host system ID %s", id)
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: id,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hs, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("could not find host system with id: %s: %s", id, err)
	}

	log.Printf("[DEBUG] Host system found: %s", hs.Reference().Value)
	return hs.(*object.HostSystem), nil
}
