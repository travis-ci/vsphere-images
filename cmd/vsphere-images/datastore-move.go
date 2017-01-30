package main

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var datastoreMoveCommand = cli.Command{
	Name:      "datastore-move",
	Usage:     "Move the image on the underlying datastore",
	ArgsUsage: "image-inventory-path src-datastore-path dest-datastore-path",
	Action:    datastoreMoveAction,
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
	},
}

func datastoreMoveAction(c *cli.Context) error {
	vSphereStringURL := c.String("vsphere-url")
	if vSphereStringURL == "" {
		return errors.New("the 'vsphere-url' flag is required")
	}

	vSphereURL, err := url.Parse(vSphereStringURL)
	if err != nil {
		return errors.Wrap(err, "parsing vSphere URL failed")
	}

	if c.NArg() != 3 {
		return errors.New("required arguments: image-inventory-path src-datastore-path dest-datastore-path")
	}

	imagePath := c.Args().Get(0)
	srcDatastorePath := c.Args().Get(1)
	destDatastorePath := c.Args().Get(2)
	ctx := context.Background()
	logger := newProgressLogger("Moving image… ")

	err = vsphereimages.DatastoreMoveImage(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), imagePath, srcDatastorePath, destDatastorePath, logger)
	if err != nil {
		return errors.Wrap(err, "moving image failed")
	}
	logger.Wait()

	return nil
}
