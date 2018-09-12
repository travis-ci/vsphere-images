package vsphereimages

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"net/url"
)

// ListImages returns a list of all of the virtual machines in a folder.
func ListImages(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, folderPath string) ([]*object.VirtualMachine, error) {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return nil, errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	vms, err := finder.VirtualMachineList(ctx, folderPath+"/*")
	if err != nil {
		return nil, errors.Wrap(err, "finding the VMs failed")
	}

	return vms, nil
}
