package vsphereimages

import (
	"context"
	"strings"
	"testing"
)

func TestMoveImageTopLevel(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	logger := newProgressLogger()
	if err = MoveImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", "/DC0/vm", "my_renamed_vm", logger); err != nil {
		t.Fatal(err)
	}
	logger.Wait()
}

func TestMoveImageIntoFolder(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	if err = createFolder(ctx, service, "/DC0/vm", "test_folder"); err != nil {
		t.Fatal(err)
	}

	logger := newProgressLogger()
	if err = MoveImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", "/DC0/vm/test_folder", "my_renamed_vm", logger); err != nil {
		t.Fatal(err)
	}
	logger.Wait()
}

func TestMoveImageIntoNonexistentFolder(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	logger := newProgressLogger()
	err = MoveImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", "/DC0/vm/test_folder", "my_renamed_vm", logger)
	if err == nil {
		t.Fatal("expected error moving image, but none occurred")
	}
	if !strings.Contains(err.Error(), "finding the destination folder failed") {
		t.Fatal(err)
	}
	logger.Wait()
}

func TestMoveImageThatDoesNotExist(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	logger := newProgressLogger()
	defer logger.Wait()
	err = MoveImage(ctx, service.URL(), false, "/DC0/vm/nonexistent_vm", "/DC0/vm", "my_renamed_vm", logger)

	if err == nil {
		t.Fatal("expected error moving image, but none occurred")
	}
	if !strings.Contains(err.Error(), "finding the VM failed") {
		t.Fatal(err)
	}
}

func TestMoveImageWithoutRename(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	ctx := context.TODO()
	if err = createFolder(ctx, service, "/DC0/vm", "test_folder"); err != nil {
		t.Fatal(err)
	}

	logger := newProgressLogger()
	if err = MoveImage(ctx, service.URL(), false, "/DC0/vm/DC0_H0_VM0", "/DC0/vm/test_folder", "DC0_H0_VM0", logger); err != nil {
		t.Fatal(err)
	}
	logger.Wait()
}

func createFolder(ctx context.Context, service *SimulatedService, location string, name string) error {
	finder, err := service.NewFinder(ctx)
	if err != nil {
		return err
	}

	folder, err := finder.Folder(ctx, location)
	if err != nil {
		return err
	}

	if _, err = folder.CreateFolder(ctx, name); err != nil {
		return err
	}

	return nil
}
