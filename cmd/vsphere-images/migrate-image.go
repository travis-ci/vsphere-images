package main

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var migrateImageCommand = cli.Command{
	Name:      "migrate-image",
	Usage:     "Migrates a VM to a new resource pool",
	ArgsUsage: "image-inventory-path",
	Action:    migrateImageAction,
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
			Name:  "pool",
			Usage: "The inventory path of the resource pool to migrate the VM to",
		},
	},
}

func migrateImageAction(c *cli.Context) error {
	vSphereStringURL := c.String("vsphere-url")
	if vSphereStringURL == "" {
		return errors.New("the 'vsphere-url' flag is required")
	}

	vSphereURL, err := url.Parse(vSphereStringURL)
	if err != nil {
		return errors.Wrap(err, "parsing vSphere URL failed")
	}

	imagePath := c.Args().Get(0)
	if imagePath == "" {
		return errors.New("image inventory path is required")
	}

	poolPath := c.String("pool")
	if poolPath == "" {
		return errors.New("pool path is required")
	}

	ctx := context.Background()
	logger := newProgressLogger("Migrating image… ")
	err = vsphereimages.MigrateImage(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), imagePath, poolPath, logger)
	if err != nil {
		return errors.Wrap(err, "migrating image failed")
	}
	logger.Wait()

	return nil
}
