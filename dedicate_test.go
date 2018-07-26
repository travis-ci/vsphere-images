package vsphereimages

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/types"
	"testing"
)

func TestIsHostCheckedOut(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	result, err := IsHostCheckedOut(ctx, service.URL(), false, "/DC0/host/DC0_C0")
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected host checked out to be true, was false")
	}
}

func TestIsHostCheckedOutEmptyCluster(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	if err = createCluster(ctx, service, "/DC0/host", "dedicated"); err != nil {
		t.Fatal(err)
	}

	result, err := IsHostCheckedOut(ctx, service.URL(), false, "/DC0/host/dedicated")
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected host checked out to be false, was true")
	}
}

// func TestCheckOutHost(t *testing.T) {
// 	service, err := StartService()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer service.Stop()
//
// 	ctx := context.TODO()
// 	if err = createCluster(ctx, service, "/DC0/host", "dedicated"); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	logger := newProgressLogger()
// 	_, err = CheckOutHost(ctx, service.URL(), false, "/DC0/host/DC0_C0", "/DC0/host/dedicated", logger)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func createCluster(ctx context.Context, service *SimulatedService, location string, name string) error {
	client, err := govmomi.NewClient(ctx, service.URL(), false)
	if err != nil {
		return err
	}

	finder := find.NewFinder(client.Client, false)

	folder, err := finder.Folder(ctx, location)
	if err != nil {
		return err
	}

	if _, err = folder.CreateCluster(ctx, name, types.ClusterConfigSpecEx{}); err != nil {
		return err
	}

	return nil
}
