package main

import (
	"context"
	"net/url"
	"path"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var copyImageCommand = cli.Command{
	Name:      "copy-image",
	Usage:     "copy image from one vCenter to another",
	ArgsUsage: "src-image-name dest-image-name",
	Action:    copyImageAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "src-url",
			Usage:  "URL to the source vCenter",
			EnvVar: "VSPHERE_IMAGES_SRC_URL",
		},
		cli.StringFlag{
			Name:   "dest-url",
			Usage:  "URL to the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_URL",
		},
		cli.BoolFlag{
			Name:   "src-insecure-skip-verify",
			Usage:  "Whether the source vCenter's certificate chain and hostname should be verified",
			EnvVar: "VSPHERE_IMAGES_SRC_INSECURE_SKIP_VERIFY",
		},
		cli.BoolFlag{
			Name:   "dest-insecure-skip-verify",
			Usage:  "Whether the destination vCenter's certificate chain and hostname should be verified",
			EnvVar: "VSPHERE_IMAGES_DEST_INSECURE_SKIP_VERIFY",
		},
		cli.StringFlag{
			Name:   "dest-sha1-fingerprint",
			Usage:  "The SHA-1 fingerprint of the TLS certificate on the destination vCenter. Format should be :-separated hexadecimal numbers. Leave empty in order to compute the fingerprint on this machine.",
			EnvVar: "VSPHERE_IMAGES_DEST_SHA1_FINGERPRINT",
		},
		cli.StringFlag{
			Name:   "dest-datastore-path",
			Usage:  "The inventory path to the datastore to put the VM in in the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_DATASTORE_PATH",
		},
		cli.StringFlag{
			Name:   "dest-pool-path",
			Usage:  "The inventory path to the resource pool to put the VM in in the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_POOL_PATH",
		},
		cli.StringFlag{
			Name:   "dest-host-path",
			Usage:  "The inventory path to the host to put the VM in in the destination vCenter",
			EnvVar: "VSPHERE_IMAGES_DEST_HOST_PATH",
		},
		cli.StringFlag{
			Name:   "dest-network-name",
			Usage:  "The name of the network to connect the copied VM to.",
			EnvVar: "VSPHERE_IMAGES_DEST_NETWORK_NAME",
		},
	},
}

func copyImageAction(c *cli.Context) error {
	srcStringURL := c.String("src-url")
	destStringURL := c.String("dest-url")

	srcURL, err := url.Parse(srcStringURL)
	if err != nil {
		return errors.Wrap(err, "parsing source URL failed")
	}
	destURL, err := url.Parse(destStringURL)
	if err != nil {
		return errors.Wrap(err, "parsing destination URL failed")
	}

	ctx := context.Background()

	source := vsphereimages.ImageSource{
		VSphereEndpoint:           srcURL,
		VSphereInsecureSkipVerify: c.Bool("src-insecure-skip-verify"),
		VMPath: c.Args().Get(0),
	}

	destination := vsphereimages.ImageDestination{
		VSphereEndpoint:           destURL,
		VSphereInsecureSkipVerify: c.Bool("dest-insecure-skip-verify"),
		VSphereSHA1Fingerprint:    c.String("dest-sha1-fingerprint"),
		FolderPath:                path.Dir(c.Args().Get(1)),
		DatastorePath:             c.String("dest-datastore-path"),
		ResourcePoolPath:          c.String("dest-pool-path"),
		HostPath:                  c.String("dest-host-path"),
		VMName:                    path.Base(c.Args().Get(1)),
		NetworkPath:               c.String("dest-network-name"),
	}

	logger := newProgressLogger("Copying image… ")
	err = vsphereimages.CopyImage(ctx, source, destination, logger)
	if err != nil {
		return errors.Wrap(err, "copying image failed")
	}
	logger.Wait()

	return nil
}
