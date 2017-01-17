ROOT_PACKAGE := github.com/travis-ci/vsphere-images
MAIN_PACKAGE := $(ROOT_PACKAGE)/cmd/vsphere-images
TEST_PACKAGES := $(ROOT_PACKAGE) $(MAIN_PACKAGE)
COVER_PACKAGES := $(ROOT_PACKAGE),$(MAIN_PACKAGE)

VERSION_VAR := main.VersionString
VERSION_VALUE ?= $(shell git describe --always --dirty --tags 2>/dev/null)
REV_VAR := main.RevisionString
REV_VALUE ?= $(shell git rev-parse HEAD 2>/dev/null || echo "???")
REV_URL_VAR := main.RevisionURLString
REV_URL_VALUE ?= https://github.com/travis-ci/vsphere-images/tree/$(shell git rev-parse HEAD 2>/dev/null || echo "'???'")
GENERATED_VAR := main.GeneratedString
GENERATED_VALUE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%S%z')
COPYRIGHT_VAR := main.CopyrightString
COPYRIGHT_VALUE ?= $(shell grep -i ^copyright LICENSE | sed 's/^[Cc]opyright //')

GOPATH := $(shell echo $${GOPATH%%:*})
GOBUILD_LDFLAGS ?= \
    -X '$(VERSION_VAR)=$(VERSION_VALUE)' \
    -X '$(REV_VAR)=$(REV_VALUE)' \
    -X '$(REV_URL_VAR)=$(REV_URL_VALUE)' \
    -X '$(GENERATED_VAR)=$(GENERATED_VALUE)' \
    -X '$(COPYRIGHT_VAR)=$(COPYRIGHT_VALUE)'

.PHONY: all
all: clean test build

.PHONY: clean
clean:
	$(RM) $(GOPATH)/bin/vsphere-images
	$(RM) -rv ./build
	find $(GOPATH)/pkg -wholename "*$(ROOT_PACKAGE)*.a" -delete

.PHONY: test
test:
	for package in $(TEST_PACKAGES); do \
		go test -x -v \
			$${package}; \
	done

.PHONY: build
build: deps
	go install -x -ldflags "$(GOBUILD_LDFLAGS)" $(MAIN_PACKAGE)

.PHONY: crossbuild
crossbuild: deps
	GOARCH=amd64 GOOS=darwin go build -o build/darwin/amd64/vsphere-images \
		-ldflags "$(GOBUILD_LDFLAGS)" $(MAIN_PACKAGE)
	GOARCH=amd64 GOOS=linux go build -o build/linux/amd64/vsphere-images \
		-ldflags "$(GOBUILD_LDFLAGS)" $(MAIN_PACKAGE)

.PHONY: distclean
distclean:
	$(RM) vendor/.deps-fetched

.PHONY: deps
deps: vendor/.deps-fetched

.PHONY: prereqs
prereqs:
	go get -u github.com/FiloSottile/gvt
	go get -u github.com/wadey/gocovmerge

.PHONY: copyright
copyright:
	sed -i "s/^Copyright.*Travis CI/Copyright © $(shell date +%Y) Travis CI/" LICENSE

vendor/.deps-fetched:
	gvt rebuild
	touch $@
