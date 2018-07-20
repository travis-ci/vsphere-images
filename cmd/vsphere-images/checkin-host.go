package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
)

var checkinHostCommand = cli.Command{
	Name:      "checkin-host",
	Usage:     "Puts the only host in the cluster into maintenance mode",
	ArgsUsage: "cluster-path",
	Action:    checkinHostAction,
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
	},
}

func checkinHostAction(c *cli.Context) error {
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

	destinationClusterPath := c.String("dest-pool")
	if destinationClusterPath == "" {
		return errors.New("destination cluster path is required")
	}

	ctx := context.Background()
	logger := newProgressLogger("Checking in host… ")
	host, err := vsphereimages.CheckInHost(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), clusterInventoryPath, destinationClusterPath, logger)
	if err != nil {
		return errors.Wrap(err, "checking in host failed")
	}
	logger.Wait()

	fmt.Println("Checked in host", host.Name())

	return nil
}
