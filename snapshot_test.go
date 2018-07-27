package vsphereimages

import (
	"context"
	"testing"
)

func TestSnapshotImage(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	logger := newProgressLogger()
	defer logger.Wait()
	if err = SnapshotImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", logger); err != nil {
		t.Fatal(err)
	}

	err = hasBaseSnapshot(ctx, service, "/DC0/vm/DC0_H0_VM0")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotImageTwice(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	logger := newProgressLogger()
	defer logger.Wait()
	if err = SnapshotImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", logger); err != nil {
		t.Fatal(err)
	}
	if err = SnapshotImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", logger); err != nil {
		t.Fatal(err)
	}

	err = hasBaseSnapshot(ctx, service, "/DC0/vm/DC0_H0_VM0")
	if err != nil {
		t.Fatal(err)
	}

	// TODO: test if this snapshot is different than the first one
}

func hasBaseSnapshot(ctx context.Context, service *SimulatedService, vmPath string) error {
	finder, err := service.NewFinder(ctx)
	if err != nil {
		return err
	}

	vm, err := finder.VirtualMachine(ctx, vmPath)
	if err != nil {
		return err
	}

	_, err = vm.FindSnapshot(ctx, "base")
	return err
}
