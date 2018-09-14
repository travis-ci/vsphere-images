package vsphereimages

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/types"
)

// RestoreBackup clones a backed up VM image into its original location within the same datacenter.
//
// The VM will be cloned initialially with its name suffixed with "-restoring" in the destination folder.
// Once the clone is complete, the cloned VM will receive a new base snapshot.
// If there is an existing VM at the destination path, it will be renamed with a "-old" suffix.
// Finally, the restored VM will be renamed to match its original name.
func RestoreBackup(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, sourceImagePath string, destinationFolderPath string, defaultDatastorePath string, defaultResourcePool string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	sourceImage, err := finder.VirtualMachine(ctx, sourceImagePath)
	if err != nil {
		return errors.Wrap(err, "finding the backup VM failed")
	}
	name := sourceImage.Name()

	destFolder, err := finder.Folder(ctx, destinationFolderPath)
	if err != nil {
		return errors.Wrap(err, "finding the destination folder failed")
	}

	var datastore *object.Datastore
	var pool *object.ResourcePool

	existingImagePath := destFolder.InventoryPath + "/" + name
	existingImage, err := finder.VirtualMachine(ctx, existingImagePath)

	if err == nil {
		datastore, err = imageDatastore(ctx, existingImage)
		if err != nil {
			return errors.Wrap(err, "finding the existing datastore failed")
		}

		pool, err = existingImage.ResourcePool(ctx)
		if err != nil {
			return errors.Wrap(err, "finding the existing resource pool failed")
		}
	} else {
		datastore, err = finder.Datastore(ctx, defaultDatastorePath)
		if err != nil {
			return errors.Wrap(err, "finding the datastore failed")
		}

		pool, err = finder.ResourcePool(ctx, defaultResourcePool)
		if err != nil {
			return errors.Wrap(err, "finding the resource pool failed")
		}
	}
	datastoreRef := datastore.Reference()
	poolRef := pool.Reference()

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: types.VirtualMachineRelocateSpec{
			Datastore: &datastoreRef,
			Pool:      &poolRef,
		},
	}

	restoringName := name + "-restoring"
	task, err := sourceImage.Clone(ctx, destFolder, restoringName, cloneSpec)
	if err != nil {
		return errors.Wrap(err, "creating VM clone task failed")
	}

	if _, err = task.WaitForResult(ctx, s); err != nil {
		return errors.Wrap(err, "cloning VM failed")
	}

	restoringImage, err := finder.VirtualMachine(ctx, destinationFolderPath+"/"+restoringName)
	if err != nil {
		return errors.Wrap(err, "finding the restoring VM failed")
	}

	if err = resnapshotImage(ctx, restoringImage, s); err != nil {
		return err // this error is already distinct enough
	}

	if existingImage != nil {
		if err = renameImage(ctx, existingImage, name+"-old", s); err != nil {
			return errors.Wrap(err, "renaming existing VM failed")
		}
	}

	if err = renameImage(ctx, restoringImage, name, s); err != nil {
		return errors.Wrap(err, "renaming restoring VM failed")
	}

	return nil
}

func resnapshotImage(ctx context.Context, vm *object.VirtualMachine, s progress.Sinker) error {
	task, err := vm.RemoveAllSnapshot(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "creating remove snapshot task failed")
	}

	if _, err = task.WaitForResult(ctx, s); err != nil {
		return errors.Wrap(err, "removing all snapshots failed")
	}

	task, err = vm.CreateSnapshot(ctx, "base", "", false, false)
	if err != nil {
		return errors.Wrap(err, "creating task to create snapshot failed")
	}

	if _, err = task.WaitForResult(ctx, s); err != nil {
		return errors.Wrap(err, "creating snapshot failed")
	}

	return nil
}

func renameImage(ctx context.Context, vm *object.VirtualMachine, newName string, s progress.Sinker) error {
	configSpec := types.VirtualMachineConfigSpec{
		Name: newName,
	}

	task, err := vm.Reconfigure(ctx, configSpec)
	if err != nil {
		return errors.Wrap(err, "creating VM rename task failed")
	}

	if _, err = task.WaitForResult(ctx, s); err != nil {
		return errors.Wrap(err, "renaming VM failed")
	}

	return nil
}

func imageDatastore(ctx context.Context, vm *object.VirtualMachine) (*object.Datastore, error) {
	var mvm mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"datastore"}, &mvm); err != nil {
		return nil, err
	}

	if len(mvm.Datastore) < 1 {
		return nil, errors.New("expected VM to have at least one datastore")
	}

	return object.NewDatastore(vm.Client(), mvm.Datastore[0]), nil
}
