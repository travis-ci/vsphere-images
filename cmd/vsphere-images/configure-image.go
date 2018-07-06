package main

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
	"github.com/vmware/govmomi/vim25/types"
)

var configureImageCommand = cli.Command{
	Name:      "configure-image",
	Usage:     "Reconfigures properties of a virtual machine",
	ArgsUsage: "image-inventory-path",
	Action:    configureImageAction,
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
		cli.IntFlag{
			Name:  "num-cpus",
			Usage: "Number of CPUs the machine should have",
		},
		cli.IntFlag{
			Name:  "cores-per-socket",
			Usage: "Number of CPU cores that each CPU should have",
		},
		cli.Int64Flag{
			Name:  "ram",
			Usage: "Amount of RAM the machine should have (in MB)",
		},
		cli.StringFlag{
			Name:  "network",
			Usage: "The name of the network to use",
		},
	},
}

func configureImageAction(c *cli.Context) error {
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

	var configSpec types.VirtualMachineConfigSpec

	numCPUs := c.Int("num-cpus")
	if numCPUs > 0 {
		configSpec.NumCPUs = int32(numCPUs)
	}

	coresPerSocket := c.Int("cores-per-socket")
	if coresPerSocket > 0 {
		configSpec.NumCoresPerSocket = int32(coresPerSocket)
	}

	ram := c.Int64("ram")
	if ram > 0 {
		configSpec.MemoryMB = ram
	}

	network := c.String("network")

	ctx := context.Background()
	logger := newProgressLogger("Configuring image… ")
	err = vsphereimages.ConfigureImage(ctx, vSphereURL, c.Bool("vsphere-insecure-skip-verify"), imagePath, configSpec, network, logger)
	if err != nil {
		return errors.Wrap(err, "configuring image failed")
	}
	logger.Wait()

	return nil
}
