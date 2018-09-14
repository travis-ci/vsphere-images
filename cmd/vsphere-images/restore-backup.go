package main

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var restoreBackupCommand = cli.Command{
	Name:      "restore-backup",
	Usage:     "restore an image from a backup clone",
	ArgsUsage: "image-path",
	Action:    restoreBackupAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "vsphere-url",
			Usage:  "URL to the vSphere SDK API",
			EnvVar: "VSPHERE_IMAGES_VSPHERE_URL",
		},
		cli.BoolFlag{
			Name:   "vsphere-insecure-skip-verify",
			Usage:  "Whether the vCenter's certificate chain and hostname should be verified",
			EnvVar: "VSPHERE_IMAGES_VSPHERE_INSECURE_SKIP_VERIFY",
		},
		cli.StringFlag{
			Name:   "datastore-path",
			Usage:  "The inventory path to the datastore to put the VM in in the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DATASTORE_PATH",
		},
		cli.StringFlag{
			Name:   "pool-path",
			Usage:  "The inventory path to the resource pool to put the VM in in the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_POOL_PATH",
		},
		cli.StringFlag{
			Name:  "dest-folder-path",
			Usage: "The inventory path to the folder the VM should be restored to",
		},
	},
}

func restoreBackupAction(c *cli.Context) error {
	ctx := context.Background()

	vSphereURL, err := url.Parse(c.String("vsphere-url"))
	if err != nil {
		return errors.Wrap(err, "parsing source URL failed")
	}

	insecure := c.Bool("vsphere-insecure-skip-verify")
	sourceImagePath := c.Args().Get(0)
	destFolderPath := c.String("dest-folder-path")
	datastorePath := c.String("datastore-path")
	poolPath := c.String("pool-path")

	logger := newProgressLogger("Restoring backup image… ")
	if err = vsphereimages.RestoreBackup(ctx, vSphereURL, insecure, sourceImagePath, destFolderPath, datastorePath, poolPath, logger); err != nil {
		return errors.Wrap(err, "restoring backup image failed")
	}

	logger.Wait()
	return nil
}
