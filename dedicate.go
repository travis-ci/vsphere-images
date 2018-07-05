package vsphereimages

import (
	"context"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/progress"
)

func CheckOutHost(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, clusterInventoryPath string, destinationClusterPath string, s progress.Sinker) (*object.HostSystem, error) {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return nil, errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	alreadyCheckedOut, err := hasCheckedOutHost(ctx, destinationClusterPath, finder)
	if err != nil {
		return nil, errors.Wrap(err, "could not determine if a host was already checked out to destination cluster")
	}

	if alreadyCheckedOut {
		return nil, errors.New("a host is already checked out to the cluster at " + destinationClusterPath)
	}

	hosts, err := finder.HostSystemList(ctx, clusterInventoryPath)
	if err != nil {
		return nil, errors.Wrap(err, "finding hosts for cluster failed")
	}

	if len(hosts) < 1 {
		return nil, errors.New("no hosts found in cluster")
	}

	var chosenHost *object.HostSystem
	for _, host := range hosts {
		canCheckOut, err := canCheckOutHost(ctx, host, finder)
		if canCheckOut {
			chosenHost = host
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed determining if host could be checked out")
		}
	}

	task, err := chosenHost.EnterMaintenanceMode(ctx, 0, true, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating the maintenance mode task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "moving the host into maintenance mode failed")
	}

	return chosenHost, nil
}

func CheckInHost(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, clusterInventoryPath string, destinationClusterPath string, s progress.Sinker) (*object.HostSystem, error) {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return nil, errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	hosts, err := finder.HostSystemList(ctx, clusterInventoryPath)
	if err != nil {
		return nil, errors.Wrap(err, "finding hosts for cluster failed")
	}

	if len(hosts) < 1 {
		return nil, errors.New("no hosts found in cluster")
	}

	if len(hosts) > 1 {
		return nil, errors.New("more than 1 host found in cluster (maybe this is the wrong cluster?)")
	}

	task, err := hosts[0].EnterMaintenanceMode(ctx, 0, true, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating the maintenance mode task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "moving the host into maintenance mode failed")
	}

	return hosts[0], nil
}

func hasCheckedOutHost(ctx context.Context, clusterPath string, finder *find.Finder) (bool, error) {
	hosts, err := finder.HostSystemList(ctx, clusterPath)
	if err != nil {
		return false, err
	}

	return len(hosts) > 0, nil
}

func canCheckOutHost(ctx context.Context, host *object.HostSystem, finder *find.Finder) (bool, error) {
	vms, err := finder.VirtualMachineList(ctx, host.InventoryPath+"/*")
	if err != nil {
		return false, err
	}

	for _, vm := range vms {
		// if any VMs on the host are not a build VM, we shouldn't check out that host
		if !isBuildVMName(vm.Name()) {
			return false, nil
		}
	}

	return true, nil
}

var buildVMNameRegexp = regexp.MustCompile("^[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}$")

func isBuildVMName(name string) bool {
	return buildVMNameRegexp.MatchString(name)
}
