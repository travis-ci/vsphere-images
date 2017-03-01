# vsphere-images

## Usage

### Copy image

Copy `foobar` from vSphere 192.0.2.1 to 192.0.2.2:

```
$ vphere-images copy-image \
	--src-url=https://admin:password@192.0.2.1/sdk \
	--dest-url=https://admin:password@192.0.2.2/sdk \
	--dest-datastore-path=/Datacenter-2/datastore/DS-1 \
	--dest-pool-path=/Datacenter-2/host/main_pool \
	--dest-host-path=/Datacenter-2/host/main_pool/host01 \
	--dest-network-name=/Datacenter-2/network/PortGroupName \
	"/Datacenter-1/vm/base/foobar" \
	"/Datacenter-2/vm/base/foobar"
```

The destination pool and host are both required, even if the destination pool has DRS enabled.

### Move image in datastore

```
$ vsphere-images datastore-move \
	--vsphere-url=https://admin:password@192.0.2.1/sdk \
	"/Datacenter-2/vm/base/foobar" \
	"[DS-1] /foobar" \
	"[DS-1] /images/foobar"
```

### Resnapshot an image

```
$ vsphere-images resnapshot \
	--vsphere-url=https://admin:password@192.0.2.1/sdk \
	"/Datacenter-2/vm/base/foobar"
```

## License

See LICENSE file.

© 2016 Travis CI GmbH
