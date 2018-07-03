package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var checkoutHostCommand = cli.Command{
	Name:      "checkout-host",
	Usage:     "Chooses a host from a cluster to bring down to maintenance mode",
	ArgsUsage: "cluster-path",
	Action:    checkoutHostAction,
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

func checkoutHostAction(c *cli.Context) error {
	vSphereStringURL := c.String("vsphere-url")
	if vSphereStringURL == "" {
		return errors.New("the 'vsphere-url' flag is required")
	}

	vSphereURL, err := url.Parse(vSphereStringURL)
	if err != nil {
		return errors.Wrap(err, "parsing vSphere URL failed")
	}

	clusterInventoryPath := c.Args().Get(0)
	if clusterInventoryPath == "" {
		return errors.New("cluster inventory path is required")
	}

	ctx := context.Background()
	logger := newProgressLogger("Checking out host… ")
	host, err := vsphereimages.CheckOutHost(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), clusterInventoryPath, logger)
	if err != nil {
		return errors.Wrap(err, "checking out host failed")
	}
	logger.Wait()

	if host == nil {
		fmt.Println("No suitable host to check out was found")
	} else {
		fmt.Println("Checked out host", host.Name())
		fmt.Println("Please move it to the desired cluster manually using the vCenter client.")
	}

	return nil
}
