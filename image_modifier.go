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

func ConfigureImage(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, imageInventoryPath string, config types.VirtualMachineConfigSpec, networkName string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	vm, err := finder.VirtualMachine(ctx, imageInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the VM failed")
	}

	if networkName != "" {
		devices, err := vm.Device(ctx)
		if err != nil {
			return errors.Wrap(err, "loading VM devices failed")
		}

		net := devices.Find("ethernet-0")
		if net == nil {
			return errors.New("could not find 'ethernet-0' device")
		}

		network, err := finder.Network(ctx, networkName)
		if err != nil {
			return errors.Wrap(err, "finding the network failed")
		}

		backing, err := network.EthernetCardBackingInfo(ctx)
		if err != nil {
			return errors.Wrap(err, "creating network backing info failed")
		}

		card := net.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
		card.Backing = backing

		config.DeviceChange = []types.BaseVirtualDeviceConfigSpec{
			&types.VirtualDeviceConfigSpec{
				Device:    net,
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
			},
		}
	}

	task, err := vm.Reconfigure(ctx, config)
	if err != nil {
		return errors.Wrap(err, "creating the VM config task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	return errors.Wrap(err, "reconfiguring the VM failed")
}

func MigrateImage(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, imageInventoryPath string, poolInventoryPath string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	vm, err := finder.VirtualMachine(ctx, imageInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the VM failed")
	}

	pool, err := finder.ResourcePool(ctx, poolInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the resource pool failed")
	}

	task, err := vm.Migrate(ctx, pool, nil, types.VirtualMachineMovePriorityDefaultPriority, types.VirtualMachinePowerStatePoweredOff)
	if err != nil {
		return errors.Wrap(err, "creating the migrate task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	return errors.Wrap(err, "migrating the VM failed")
}
