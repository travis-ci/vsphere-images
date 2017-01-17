# vsphere-images

## Usage

### Copy image

Copy `foobar` from vSphere 192.0.2.1 to 192.0.2.2:

```
$ vphere-images copy-image \
	-src-url=https://admin:password@192.0.2.1/sdk \
	-dest-url=https://admin:password@192.0.2.2/sdk \
	-dest-datastore-path=/Datacenter-2/datastore/DS-1 \
	-dest-pool-path=/Datacenter-2/host/main_pool \
	-dest-host-path=/Datacenter-2/host/main_pool/host01 \
	"/Datacenter-1/vm/base/foobar" \
	"/Datacenter-2/vm/base/foobar"
```

## License

See LICENSE file.

© 2016 Travis CI GmbH
