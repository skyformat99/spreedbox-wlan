PWD := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

GOPKG = golang.struktur.de/spreedbox/spreedbox-wlan
GOPATH = "$(CURDIR)/vendor:$(CURDIR)"
SYSTEM_GOPATH = /usr/share/gocode/src/

DIST := $(PWD)/dist
DIST_SRC := $(DIST)/src

all:

$(DIST_SRC):
	mkdir -p $@

dist_gopath: $(DIST_SRC)
	if [ -d "$(SYSTEM_GOPATH)" ]; then find $(SYSTEM_GOPATH) -mindepth 1 -maxdepth 1 -type d \
		-exec ln -sf {} $(DIST_SRC) \; ; fi
	if [ ! -d "$(SYSTEM_GOPATH)" ]; then find $(CURDIR)/vendor/src -mindepth 1 -maxdepth 1 -type d \
		-exec ln -sf {} $(DIST_SRC) \; ; fi

goget:
	if [ -z "$(DEB_BUILDING)" ]; then GOPATH=$(GOPATH) go get launchpad.net/godeps; fi
	if [ -z "$(DEB_BUILDING)" ]; then GOPATH=$(GOPATH) $(CURDIR)/vendor/bin/godeps -u dependencies.tsv; fi
	mkdir -p $(shell dirname "$(CURDIR)/vendor/src/$(GOPKG)")
	ln -sf $(PWD) $(shell dirname "$(CURDIR)/vendor/src/$(GOPKG)")

build: goget
	GOPATH=$(GOPATH) go build $(shell find wlan* -type d -exec echo ./{} \;)

test: goget
	GOPATH=$(GOPATH) go test $(shell find wlan* -type d -exec echo ./{} \;)

.PHONY: all dist_gopath goget build
