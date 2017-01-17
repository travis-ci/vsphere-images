package vsphereimages

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/types"
)

type ImageSource struct {
	// VSphereEndpoint is the URL to the vSphere API, including the
	// credentials. Usually this is something like
	// `https://user@password:your-vsphere-host/sdk`.
	VSphereEndpoint *url.URL

	// VSphereInsecureSkipVerify controls whether the server's certificate
	// chain and hostname should be verified. If VSphereInsecureSkipVerify is
	// true, any certificate presented by the server and any host name in that
	// certificate will be accepted. See also
	// crypto/tls.Config.InsecureSkipVerify.
	VSphereInsecureSkipVerify bool

	// VMPath is the inventory path to the source VM. This is usually something
	// like `/your-datacenter/vm/folder-name/vm-name`.
	VMPath string
}

type ImageDestination struct {
	// VSphereEndpoint is the URL to the vSphere API, including the
	// credentials. Usually this is something like
	// `https://user@password:your-vsphere-host/sdk`.
	VSphereEndpoint *url.URL

	// VSphereInsecureSkipVerify controls whether the server's certificate
	// chain and hostname should be verified. If VSphereInsecureSkipVerify is
	// true, any certificate presented by the server and any host name in that
	// certificate will be accepted. See also
	// crypto/tls.Config.InsecureSkipVerify.
	VSphereInsecureSkipVerify bool

	// VSphereSHA1Fingerprint is the SHA-1 fingerprint of the vSphere API
	// endpoint. It should be formatted as a series of :-separated uppercase
	// hexadecimal numbers. If the string is empty, the ImageCopier will
	// attempt to connect to the endpoint and use that fingerprint.
	VSphereSHA1Fingerprint string

	// FolderPath is the inventory path to the folder in the destination
	// vCenter to copy the VM to. Folder inventory paths usually look something
	// like `/your-datacenter/vm/folder-name`.
	FolderPath string

	// DatastorePath is the inventory path to the datastore in the destination
	// vCenter to copy the VM to. Datastore inventory paths usually look
	// something like `/your-datacenter/datastore/datastore-name`.
	DatastorePath string

	// ResourcePoolPath is the inventory path to the resource pool in the
	// destination vCenter to copy the VM to. Resource pool paths usually look
	// something like `/your-datacenter/host/pool-name`.
	ResourcePoolPath string

	// HostPath is the inventory path to the host in the destination vCenter to
	// copy the VM to. Note that this host should be a part of the resource
	// pool in ResourcePoolPath. Host inventory paths for hosts in a resource
	// pool usually look something like
	// `/your-datacenter/host/pool-name/host-name`.
	HostPath string

	// VMName is the name to give to the destination VM.
	VMName string
}

func CopyImage(ctx context.Context, source ImageSource, destination ImageDestination) error {
	srcClient, err := govmomi.NewClient(ctx, source.VSphereEndpoint, source.VSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating source vSphere client failed")
	}

	destClient, err := govmomi.NewClient(ctx, destination.VSphereEndpoint, destination.VSphereInsecureSkipVerify)
	if err != nil {
		return errors.Wrap(err, "creating destination vSphere client failed")
	}

	srcFinder := find.NewFinder(srcClient.Client, false)
	destFinder := find.NewFinder(destClient.Client, false)

	srcVM, err := srcFinder.VirtualMachine(ctx, source.VMPath)
	if err != nil {
		return errors.Wrap(err, "finding the source VM failed")
	}

	destFolder, err := destFinder.Folder(ctx, destination.FolderPath)
	if err != nil {
		return errors.Wrap(err, "finding the destination folder failed")
	}
	destFolderRef := destFolder.Reference()

	destDatastore, err := destFinder.Datastore(ctx, destination.DatastorePath)
	if err != nil {
		return errors.Wrap(err, "finding the destination datastore failed")
	}
	destDatastoreRef := destDatastore.Reference()

	destPool, err := destFinder.ResourcePool(ctx, destination.ResourcePoolPath)
	if err != nil {
		return errors.Wrap(err, "finding the destination resource pool failed")
	}
	destPoolRef := destPool.Reference()

	destHost, err := destFinder.HostSystem(ctx, destination.HostPath)
	if err != nil {
		return errors.Wrap(err, "finding the destination host failed")
	}
	destHostRef := destHost.Reference()

	if destination.VSphereSHA1Fingerprint == "" {
		hostPort := destination.VSphereEndpoint.Host
		if !hasPort(hostPort) {
			hostPort = hostPort + ":" + portMap[destination.VSphereEndpoint.Scheme]
		}

		destination.VSphereSHA1Fingerprint, err = findSHA1Fingerprint(hostPort)
		if err != nil {
			return errors.Wrap(err, "finding the SHA-1 fingerprint of the destination vCenter failed")
		}
	}

	if destination.VSphereEndpoint.User == nil {
		return errors.New("destination vSphere endpoint doesn't have username and password set")
	}
	username := destination.VSphereEndpoint.User.Username()
	password, passwordSet := destination.VSphereEndpoint.User.Password()
	if !passwordSet {
		return errors.New("destination vSphere endpoint doesn't have password set")
	}

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: types.VirtualMachineRelocateSpec{
			Service: &types.ServiceLocator{
				Credential: &types.ServiceLocatorNamePassword{
					Username: username,
					Password: password,
				},
				InstanceUuid:  destClient.Client.ServiceContent.About.InstanceUuid,
				SslThumbprint: destination.VSphereSHA1Fingerprint,
				Url:           fmt.Sprintf("%s://%s", destination.VSphereEndpoint.Scheme, destination.VSphereEndpoint.Host),
			},
			Folder:    &destFolderRef,
			Datastore: &destDatastoreRef,
			Pool:      &destPoolRef,
			Host:      &destHostRef,
		},
	}

	cloneTask, err := srcVM.Clone(ctx, destFolder, destination.VMName, cloneSpec)
	if err != nil {
		return errors.Wrap(err, "creating VM clone task failed")
	}

	return errors.Wrap(cloneTask.Wait(ctx), "cloning VM failed")
}

func findSHA1Fingerprint(hostport string) (string, error) {
	conn, err := tls.Dial("tcp", hostport, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "dial failed")
	}

	err = conn.Handshake()
	if err != nil {
		return "", errors.Wrap(err, "handshake failed")
	}

	certificate := conn.ConnectionState().PeerCertificates[0]
	sha1Fingerprint := sha1.Sum(certificate.Raw)
	formattedFingerprint := make([]string, 0, len(sha1Fingerprint))
	for _, b := range sha1Fingerprint {
		formattedFingerprint = append(formattedFingerprint, fmt.Sprintf("%02X", b))
	}
	return strings.Join(formattedFingerprint, ":"), nil
}

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
//
// Copied from Go 1.7 net/http: https://github.com/golang/go/blob/230a376b5a67f0e9341e1fa47e670ff762213c83/src/net/http/http.go
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }
