package main

import (
	"context"
	"net/url"
	"path"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var moveImageCommand = cli.Command{
	Name:      "move-image",
	Usage:     "Renames and moves a VM within a single vCenter instance",
	ArgsUsage: "image-inventory-path",
	Action:    moveImageAction,
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

func moveImageAction(c *cli.Context) error {
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

	destinationImagePath := c.Args().Get(1)
	if destinationImagePath == "" {
		return errors.New("destination image inventory path is required")
	}

	destinationFolderPath := path.Dir(destinationImagePath)
	newName := path.Base(destinationImagePath)

	ctx := context.Background()
	logger := newProgressLogger("Moving image… ")
	err = vsphereimages.MoveImage(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), imagePath, destinationFolderPath, newName, logger)
	if err != nil {
		return errors.Wrap(err, "moving image failed")
	}
	logger.Wait()

	return nil
}
