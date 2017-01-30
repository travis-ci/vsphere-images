package vsphereimages

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/types"
)

func DatastoreMoveImage(ctx context.Context, vSphereEndpoint *url.URL, vSphereInsecureSkipVerify bool, imageInventoryPath, srcDatastorePath, dstDatastorePath string, s progress.Sinker) error {
	client, err := govmomi.NewClient(ctx, vSphereEndpoint, vSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating vSphere client failed")
	}

	finder := find.NewFinder(client.Client, false)

	vm, err := finder.VirtualMachine(ctx, imageInventoryPath)
	if err != nil {
		return errors.Wrap(err, "finding the VM failed")
	}

	var mvm mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"datastore", "parent"}, &mvm)
	if err != nil {
		return errors.Wrap(err, "finding the VM's datastore failed")
	}

	if len(mvm.Datastore) != 1 {
		return errors.Errorf("VM was expected to have 1 datastore, but had %d", len(mvm.Datastore))
	}

	var mds mo.Datastore
	err = client.PropertyCollector().RetrieveOne(ctx, mvm.Datastore[0].Reference(), nil, &mds)
	if err != nil {
		return errors.Wrap(err, "getting information about datastore failed")
	}

	ds := object.NewDatastore(client.Client, mvm.Datastore[0])
	e, err := finder.Element(ctx, mvm.Datastore[0])
	if err != nil {
		return errors.Wrap(err, "looking up datastore path failed")
	}
	ds.InventoryPath = e.Path

	browser, err := ds.Browser(ctx)
	if err != nil {
		return errors.Wrap(err, "creating browser failed")
	}

	task, err := browser.SearchDatastore(ctx, ds.Path(srcDatastorePath), &types.HostDatastoreBrowserSearchSpec{
		MatchPattern: []string{"*.vmx"},
	})
	if err != nil {
		return errors.Wrap(err, "creating task to search for VM config file failed")
	}

	taskInfo, err := task.WaitForResult(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "searching for VM config file failed")
	}

	var results []types.HostDatastoreBrowserSearchResults

	switch r := taskInfo.Result.(type) {
	case types.HostDatastoreBrowserSearchResults:
		results = []types.HostDatastoreBrowserSearchResults{r}
	case types.ArrayOfHostDatastoreBrowserSearchResults:
		results = r.HostDatastoreBrowserSearchResults
	default:
		panic(fmt.Sprintf("unknown result type: %T", r))
	}

	var vmxFilename string
	for _, result := range results {
		for _, f := range result.File {
			vmxFilename = f.GetFileInfo().Path
			break
		}
	}
	if vmxFilename == "" {
		return errors.New("couldn't find *.vmx file in the source path")
	}

	if mvm.Parent == nil {
		return errors.New("expected VM to have a parent, but was nil")
	}
	folder := object.NewFolder(client.Client, *mvm.Parent)

	mes, err := mo.Ancestors(ctx, client.Client, client.Client.ServiceContent.PropertyCollector, ds.Reference())
	if err != nil {
		return errors.Wrap(err, "getting datastore's ancestors to find datacenter failed")
	}

	var dc *object.Datacenter
	for _, me := range mes {
		if me.Self.Type == "Datacenter" {
			dc = object.NewDatacenter(client.Client, me.Self)
			break
		}
	}
	if dc == nil {
		return errors.New("couldn't find a datacenter as an ancestor of the datastore")
	}

	pool, err := vm.ResourcePool(ctx)
	if err != nil {
		return errors.Wrap(err, "getting the VM's resource pool failed")
	}

	src := ds.Path(srcDatastorePath)
	dst := ds.Path(dstDatastorePath)

	err = vm.Unregister(ctx)
	if err != nil {
		return errors.Wrap(err, "unregistering the VM failed")
	}

	m := object.NewFileManager(client.Client)
	task, err = m.MoveDatastoreFile(ctx, src, dc, dst, dc, false)
	if err != nil {
		return errors.Wrap(err, "creating task to move image files failed")
	}

	_, err = task.WaitForResult(ctx, s)
	if err != nil {
		return errors.Wrap(err, "moving image files in datastore failed")
	}

	task, err = folder.RegisterVM(ctx, dst+"/"+vmxFilename, "", false, pool, nil)
	if err != nil {
		return errors.Wrap(err, "creating task to register VM failed")
	}

	_, err = task.WaitForResult(ctx, s)
	return errors.Wrap(err, "registering VM failed")
}
