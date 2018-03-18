PREFIX:=/usr/local
CC:=/bin/cc
CXX:=/bin/c++

VERSION=$(shell git describe --tags)
SRCDIR=$(shell pwd)

export GOPATH=$(SRCDIR)/workdir
GOSRC=$(GOPATH)/src
GOBIN=$(GOPATH)/bin
CIELPATH=$(GOSRC)/ciel

DISTDIR=$(SRCDIR)/instdir
GLIDE=$(GOBIN)/glide

all: build

$(CIELPATH):
	mkdir -p $(DISTDIR)/bin
	mkdir -p $(DISTDIR)/libexec/ciel-plugin
	mkdir -p $(GOSRC)
	mkdir -p $(GOBIN)
	ln -f -s -T $(SRCDIR) $(CIELPATH)

$(GLIDE):
	curl -\# https://glide.sh/get | PATH=$(GOBIN):$(PATH) sh

deps: $(CIELPATH) $(GLIDE) $(SRCDIR)/glide.yaml
	cd $(CIELPATH)
	$(GLIDE) install
	cd $(SRCDIR)

config:
	cp $(SRCDIR)/_config.go $(SRCDIR)/config.go
	sed 's,__VERSION__,$(VERSION),g' -i $(SRCDIR)/config.go
	sed 's,__PREFIX__,$(PREFIX),g' -i $(SRCDIR)/config.go

$(DISTDIR)/bin/ciel: deps config
	export CC
	export CXX
	go build -o $@ ciel

$(DISTDIR)/libexec/ciel-plugin: plugin/*
	cp -fR $^ $@

build: $(DISTDIR)/bin/ciel $(DISTDIR)/libexec/ciel-plugin

clean:
	rm -rf $(GOPATH)
	rm -rf $(DISTDIR)
	rm -rf $(SRCDIR)/vendor
	git clean -f -d $(SRCDIR)

install:
	cp -R $(DISTDIR)/* $(PREFIX)

.PHONY: all deps config build install clean
