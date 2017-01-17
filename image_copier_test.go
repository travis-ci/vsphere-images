package vsphereimages_test

import (
	"context"
	"net/url"

	vsphereimages "github.com/travis-ci/vsphere-images"
)

func ExampleCopyImage() {
	vSphereSourceURL, _ := url.Parse("https://admin@password:vsphere1.example.com/sdk")
	vSphereDestinationURL, _ := url.Parse("https://admin@password:vsphere2.example.com/sdk")

	source := vsphereimages.ImageSource{
		VSphereEndpoint: vSphereSourceURL,
		VMPath:          "/dc01/vm/base vms/foo",
	}
	destination := vsphereimages.ImageDestination{
		VSphereEndpoint:  vSphereDestinationURL,
		FolderPath:       "/dc02/vm/base vms/",
		DatastorePath:    "/dc02/datastore/main-datastore",
		ResourcePoolPath: "/dc02/host/main-pool",
		HostPath:         "/dc02/host/main-pool/host01",
		VMName:           "foo",
	}

	vsphereimages.CopyImage(context.TODO(), source, destination, nil)
}
