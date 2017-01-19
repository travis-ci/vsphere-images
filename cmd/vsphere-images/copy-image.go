package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"
	vsphereimages "github.com/travis-ci/vsphere-images"
	"github.com/urfave/cli"
	"github.com/vmware/govmomi/vim25/progress"
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

type progressLogger struct {
	prefix string
	wg     sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func newProgressLogger(prefix string) *progressLogger {
	p := &progressLogger{
		prefix: prefix,

		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(tick, ch)
			if err != nil {
				stop = true
			}
		case <-p.done:
			stop = true
		case <-tick.C:
			fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		}
	}

	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		fmt.Fprintf(os.Stderr, "\r%sError: %s\n", p.prefix, err)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		fmt.Fprintf(os.Stderr, "\r%sOK\n", p.prefix)
	}
}

func (p *progressLogger) loopB(tick *time.Ticker, ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		case <-tick.C:
			line := "\r" + p.prefix
			if r != nil {
				line += fmt.Sprintf("(%.0f%%", r.Percentage())
				detail := r.Detail()
				if detail != "" {
					line += fmt.Sprintf(", %s", detail)
				}
				line += ")"
			}
			fmt.Fprintf(os.Stderr, "%s", line)
		}
	}

	return err
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}
