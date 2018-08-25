package folder

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"log"
)

func FromName(client *govmomi.Client, name string) (*object.Folder, error) {
	log.Printf("[DEBUG] Locating folder with Name %s", name)
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fd, err := finder.FolderOrDefault(ctx, name)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Folder with Name %s found", fd.Name())
	return fd, nil
}
