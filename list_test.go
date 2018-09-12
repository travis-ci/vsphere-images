package vsphereimages

import (
	"context"
	"testing"
)

func TestListImages(t *testing.T) {
	service, err := StartService()
	if err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	vms, err := ListImages(context.TODO(), service.URL(), false, "/DC0/vm")
	if err != nil {
		t.Fatal(err)
	}

	if len(vms) != 8 {
		t.Fatalf("unexpected number of vms, expected 8, got %d", len(vms))
	}

	if vms[0].Name() != "DC0_H0_VM0" {
		t.Fatalf("unexpected first VM name, got %s", vms[0].Name())
	}
}
