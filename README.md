# vsphere-images

## Usage

### Copy image

Copy `foobar` from vSphere 192.0.2.1 to 192.0.2.2:

```
$ vphere-images copy-image \
	-src=https://admin:password@192.0.2.1/sdk \
	-dest=https://admin:password@192.0.2.2/sdk \
	-dest-resource-pool=main_pool \
	-dest-network=dvPortGroup-Foo \
	-dest-datastore=DS-1 \
	-dest-folder="/Datacenter-1/vm/base" \
	-src-temp-folder="/Datacenter-1/vm/tmp" \
	-dest-vcenter-shortname="2_2" \
	"/Datacenter-1/vm/base/foobar"
```

This takes the image named `foobar` in the folder `base` in 192.0.2.1 and makes a clone of it, places it in the `dest-datastore` datastore in a folder named `2_2_templates` and registers it in 192.0.2.2 as `base/foobar`.

## License

See LICENSE file.

© 2016 Travis CI GmbH
