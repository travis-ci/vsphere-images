package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

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
		cli.StringFlag{
			Name:  "dest-pool",
			Usage: "Path to cluster where the host will be moved",
		},
		cli.BoolFlag{
			Name:  "dry-run, n",
			Usage: "If enabled, only checks if a host is checked out to the destination cluster",
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

	destinationClusterPath := c.String("dest-pool")
	if destinationClusterPath == "" {
		return errors.New("destination cluster path is required")
	}

	dryRun := c.Bool("dry-run")

	ctx := context.Background()

	if dryRun {
		checkedOut, err := vsphereimages.IsHostCheckedOut(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), destinationClusterPath)
		if err != nil {
			return errors.Wrap(err, "finding checked out host failed")
		}

		if checkedOut {
			return nil
		}

		// report a non-zero exit code to indicate there is no checked out host
		os.Exit(1)
	} else {
		clusterInventoryPath := c.Args().Get(0)
		if clusterInventoryPath == "" {
			return errors.New("cluster inventory path is required")
		}

		logger := newProgressLogger("Checking out host… ")
		host, err := vsphereimages.CheckOutHost(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), clusterInventoryPath, destinationClusterPath, logger)
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
	}

	return nil
}
