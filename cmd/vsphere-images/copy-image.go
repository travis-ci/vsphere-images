package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

var copyImageCommand = cli.Command{
	Name:      "copy-image",
	Usage:     "copy image from one vCenter to another",
	ArgsUsage: "image-name",
	Action:    copyImageAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "src",
			Usage:  "URL to the source vCenter",
			EnvVar: "VSPHERE_IMAGES_SRC_URL",
		},
		cli.StringFlag{
			Name:   "dest",
			Usage:  "URL to the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_URL",
		},
		cli.StringFlag{
			Name:   "dest-vcenter-shortname",
			Usage:  "Short name for destination vCenter to use in paths",
			EnvVar: "VSPHERE_IMAGES_DEST_VCENTER_SHORTNAME",
		},
		cli.StringFlag{
			Name:   "src-temp-folder",
			Usage:  "Inventory folder in the source vCenter to use for temporary storage",
			EnvVar: "VSPHERE_IMAGES_SRC_TEMP_FOLDER",
		},
		cli.StringFlag{
			Name:   "dest-vcenter-ssl-thumbprint",
			Usage:  "The :-separated SHA1 SSL fingerprint for the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_VCENTER_SSL_THUMBPRINT",
		},
	},
}

func copyImageAction(c *cli.Context) error {
	srcStringURL := c.String("src")
	destStringURL := c.String("dest")

	srcURL, err := url.Parse(srcStringURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse source URL: %v\n", err)
		os.Exit(1)
	}
	destURL, err := url.Parse(destStringURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse destination URL: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srcClient, err := govmomi.NewClient(ctx, srcURL, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't create source vSphere client: %v\n", err)
		os.Exit(1)
	}

	_, err = govmomi.NewClient(ctx, destURL, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't create destination vSphere client: %v\n", err)
		os.Exit(1)
	}

	name := fmt.Sprintf("%s_%s", c.String("dest-vcenter-shortname"), path.Base(c.Args().First()))

	_, err = cloneVM(ctx, srcClient, c.Args().First(), c.String("src-temp-folder"), name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cloning VM: %v\n", err)
		os.Exit(1)
	}

	//  unregister_vm
	//
	//  export GOVC_URL="${BASE_VM_DST_VCENTER}"
	//  # move the $Base_VM into a folder for $vCenter02 inside the datastore
	//  move_vm_in_datastore
	//
	//  # register the $Base_VM in $vCenter02
	//  register_vm_in_vcenter "${BASE_VM_DST_DATASTORE}"
	//
	//  # update network card
	//  update_network "${BASE_VM_DST_FOLDER}/${BASE_VM_NAME}" "${BASE_VM_NETWORK}"
	//
	//  # this will remove any existing snapshots and make a new `base`
	//  base_snapshot "${BASE_VM_DST_FOLDER}/${BASE_VM_NAME}"

	return nil
}

func cloneVM(ctx context.Context, srcClient *govmomi.Client, baseVMPath, srcTempFolder, name string) (*object.VirtualMachine, error) {
	// Find the base VM
	vm, err := findVM(ctx, srcClient, baseVMPath)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't find base VM")
	}

	searchIndex := object.NewSearchIndex(srcClient.Client)
	folderRef, err := searchIndex.FindByInventoryPath(ctx, srcTempFolder)
	if err != nil {
		return nil, errors.Wrap(err, "error searching for src temp folder")
	}
	if folderRef == nil {
		return nil, errors.Errorf("src temp folder not found: %s", srcTempFolder)
	}
	folder, ok := folderRef.(*object.Folder)
	if !ok {
		return nil, errors.Errorf("src temp folder is not a folder but a %T", folderRef)
	}

	cloneSpec := types.VirtualMachineCloneSpec{}

	cloneTask, err := vm.Clone(ctx, folder, name, cloneSpec)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't clone VM")
	}

	info, err := cloneTask.WaitForResult(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error cloning VM")
	}

	return object.NewVirtualMachine(srcClient.Client, info.Result.(types.ManagedObjectReference)), nil
}

func findVM(ctx context.Context, client *govmomi.Client, path string) (*object.VirtualMachine, error) {
	searchIndex := object.NewSearchIndex(client.Client)
	vmRef, err := searchIndex.FindByInventoryPath(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "error searching for VM")
	}
	if vmRef == nil {
		return nil, errors.Wrap(err, "couldn't find VM")
	}

	vm, ok := vmRef.(*object.VirtualMachine)
	if !ok {
		return nil, errors.Errorf("VM is not a VM but a %T", vmRef)
	}

	return vm, nil
}

func unregisterVM(ctx context.Context, client *govmomi.Client, path string) error {
	vm, err := findVM(ctx, client, path)
	if err != nil {
		return errors.Wrap(err, "couldn't find VM to unregister")
	}

	return errors.Wrap(vm.Unregister(ctx), "couldn't unregister VM")
}
