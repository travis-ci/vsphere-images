package vsphereimages

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/progress"
)

func SnapshotImage(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, imageInventoryPath string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	vm, err := finder.VirtualMachine(ctx, imageInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the VM failed")
	}

	removeTask, err := vm.RemoveAllSnapshot(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "creating task to remove all snapshots failed")
	}

	_, err = removeTask.WaitForResult(ctx, s)
	if err != nil {
		return errors.Wrap(err, "removing all snapshots failed")
	}

	createTask, err := vm.CreateSnapshot(ctx, "base", "", false, false)
	if err != nil {
		return errors.Wrap(err, "creating task to create snapshot failed")
	}

	_, err = createTask.WaitForResult(ctx, s)
	return errors.Wrap(err, "creating snapshot failed")
}
