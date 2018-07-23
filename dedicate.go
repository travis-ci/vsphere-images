package vsphereimages

import (
	"context"
	"math"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/types"
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

	cluster, err := finder.ClusterComputeResource(ctx, destinationClusterPath)
	if err != nil {
		return nil, errors.Wrap(err, "finding the destination cluster failed")
	}

	chosenHost, err := chooseAvailableHost(ctx, hosts, finder)
	if err != nil {
		return nil, err
	}

	inMaintenanceMode, err := isHostInMaintenanceMode(ctx, chosenHost)
	if err != nil {
		return nil, errors.Wrap(err, "checking if host is in maintenance mode already failed")
	}

	// if the host is already in maintenance mode, skip this but still do the remaining work
	if !inMaintenanceMode {
		task, err := chosenHost.EnterMaintenanceMode(ctx, 0, true, nil)
		if err != nil {
			return nil, errors.Wrap(err, "creating the enter maintenance mode task failed")
		}

		_, err = task.WaitForResult(ctx, s)
		if err != nil {
			return nil, errors.Wrap(err, "moving the host into maintenance mode failed")
		}
	}

	task, err := moveHostToCluster(ctx, chosenHost, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "creating the move host task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "moving host to destination cluster failed")
	}

	task, err = chosenHost.ExitMaintenanceMode(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "creating the exit maintenance mode task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "bringing the host out of maintenance mode failed")
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

	cluster, err := finder.ClusterComputeResource(ctx, destinationClusterPath)
	if err != nil {
		return nil, errors.Wrap(err, "finding the destination cluster failed")
	}

	inMaintenanceMode, err := isHostInMaintenanceMode(ctx, hosts[0])
	if err != nil {
		return nil, errors.Wrap(err, "checking if host is in maintenance mode already failed")
	}

	// if the host is already in maintenance mode, skip this but still do the remaining work
	if !inMaintenanceMode {
		task, err := hosts[0].EnterMaintenanceMode(ctx, 0, true, nil)
		if err != nil {
			return nil, errors.Wrap(err, "creating the enter maintenance mode task failed")
		}

		_, err = task.WaitForResult(ctx, s)
		if err != nil {
			return nil, errors.Wrap(err, "moving the host into maintenance mode failed")
		}
	}

	task, err := moveHostToCluster(ctx, hosts[0], cluster)
	if err != nil {
		return nil, errors.Wrap(err, "creating the move host task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "moving host to destination cluster failed")
	}

	task, err = hosts[0].ExitMaintenanceMode(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "creating the exit maintenance mode task failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return nil, errors.Wrap(err, "bringing the host out of maintenance mode failed")
	}

	return hosts[0], nil
}

func hasCheckedOutHost(ctx context.Context, clusterPath string, finder *find.Finder) (bool, error) {
	_, err := finder.HostSystemList(ctx, clusterPath)
	if _, ok := err.(*find.NotFoundError); ok {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func chooseAvailableHost(ctx context.Context, hosts []*object.HostSystem, finder *find.Finder) (*object.HostSystem, error) {
	var chosenHost *object.HostSystem
	fewestBuildVMs := math.MaxUint32
	for _, host := range hosts {
		nonBuildVMCount, buildVMCount, err := hostVMCounts(ctx, host, finder)
		if err != nil {
			return nil, errors.Wrap(err, "failed determining if host could be checked out")
		}

		if nonBuildVMCount > 0 {
			continue
		}

		if buildVMCount < fewestBuildVMs {
			chosenHost = host
			fewestBuildVMs = buildVMCount

			if buildVMCount == 0 {
				break // we don't need to keep looping if we found a host with no VMs on it
			}
		}
	}

	return chosenHost, nil
}

func hostVMCounts(ctx context.Context, host *object.HostSystem, finder *find.Finder) (nonBuildCount int, buildCount int, err error) {
	vms, err := finder.VirtualMachineList(ctx, host.InventoryPath+"/*")
	if err != nil {
		if _, ok := err.(*find.NotFoundError); ok {
			return 0, 0, nil
		}
		return
	}

	for _, vm := range vms {
		// if any VMs on the host are not a build VM, we shouldn't check out that host
		isBuildVM := isBuildVMName(vm.Name())

		if isBuildVM {
			buildCount++
		} else {
			nonBuildCount++
		}
	}

	return
}

// build VMs have names that are UUIDs, so this regexp matches a UUID
//
// UUIDs are a sequence of groups of hex digits separated by dashes.
// The count of digits in each group is:
//    8-4-4-4-12
var buildVMNameRegexp = regexp.MustCompile("^[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}$")

func isBuildVMName(name string) bool {
	return buildVMNameRegexp.MatchString(name)
}

func isHostInMaintenanceMode(ctx context.Context, host *object.HostSystem) (bool, error) {
	var mh mo.HostSystem
	if err := host.Properties(ctx, host.Reference(), []string{"runtime.inMaintenanceMode"}, &mh); err != nil {
		return false, err
	}
	return mh.Runtime.InMaintenanceMode, nil
}

func moveHostToCluster(ctx context.Context, host *object.HostSystem, destinationCluster *object.ClusterComputeResource) (*object.Task, error) {
	// The structure of this method is largely copied from existing methods from the 'object' package in govmomi.
	// The pattern of creating a task, then calling the 'methods' version of it, and returning a new task wrapping
	// the result is how other real APIs on govmomi objects are done.
	//
	// There isn't a function on ClusterComputeResource to run this task, though, so we have to do this part ourselves.

	req := types.MoveInto_Task{
		This: destinationCluster.Reference(),
		Host: []types.ManagedObjectReference{host.Reference()},
	}

	res, err := methods.MoveInto_Task(ctx, destinationCluster.Client(), &req)
	if err != nil {
		return nil, err
	}

	return object.NewTask(destinationCluster.Client(), res.Returnval), nil
}
