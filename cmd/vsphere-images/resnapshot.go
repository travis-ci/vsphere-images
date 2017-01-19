package main

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var resnapshotCommand = cli.Command{
	Name:      "resnapshot",
	Usage:     "Remove (if it exists) and add a new 'base' snapshot to an image",
	ArgsUsage: "image-inventory-path",
	Action:    resnapshotAction,
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

func resnapshotAction(c *cli.Context) error {
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

	ctx := context.Background()
	logger := newProgressLogger("Snapshotting image… ")
	err = vsphereimages.SnapshotImage(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), imagePath, logger)
	if err != nil {
		return errors.Wrap(err, "snapshotting image failed")
	}
	logger.Wait()

	return nil
}
