package vsphereimages

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/types"
)

func MoveImage(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, imageInventoryPath string, newFolderPath string, newName string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	folder, err := finder.Folder(ctx, newFolderPath)
	if err != nil {
		return errors.Wrap(err, "finding the destination folder failed")
	}

	vm, err := finder.VirtualMachine(ctx, imageInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the VM failed")
	}

	configSpec := types.VirtualMachineConfigSpec{
		Name: newName,
	}

	task, err := vm.Reconfigure(ctx, configSpec)
	if err != nil {
		return errors.Wrap(err, "creating the VM rename task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return errors.Wrap(err, "renaming the VM failed")
	}

	task, err = folder.MoveInto(ctx, []types.ManagedObjectReference{vm.Reference()})
	if err != nil {
		return errors.Wrap(err, "creating the VM move task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	return errors.Wrap(err, "moving the VM failed")
}
