package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "vSphere Image Manager"
	app.Version = VersionString
	app.Author = "Travis CI GmbH"
	app.Email = "contact+vsphere-images@travis-ci.org"

	app.Commands = []cli.Command{
		checkinHostCommand,
		checkoutHostCommand,
		copyImageCommand,
		moveImageCommand,
		configureImageCommand,
		migrateImageCommand,
		resnapshotCommand,
		datastoreMoveCommand,
		restoreBackupCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "an error occurred: %v\n", err)
		os.Exit(1)
	}
}
