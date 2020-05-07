PROJ=ciel
ORG_PATH=github.com/AOSC-Dev
REPO_PATH=$(ORG_PATH)/$(PROJ)
PREFIX:=/usr/local
CC:=/usr/bin/cc
CXX:=/usr/bin/c++

VERSION=$(shell git describe --tags)
SRCDIR=$(shell pwd)

export GOPATH=$(SRCDIR)/workdir
export CC
export CXX

ARCH:=$(shell uname -m)

GOSRC=$(GOPATH)/src
GOBIN=$(GOPATH)/bin
CIELPATH=$(GOSRC)/ciel
LD_FLAGS="-w -X $(REPO_PATH)/config.Version=$(VERSION) -X $(REPO_PATH)/config.Prefix=$(PREFIX)"
DISTDIR=$(SRCDIR)/instdir

all: build

$(CIELPATH):
	mkdir -p $(DISTDIR)/bin
	mkdir -p $(DISTDIR)/libexec/ciel-plugin
	mkdir -p $(GOSRC)
	mkdir -p $(GOBIN)
	ln -f -s -T $(SRCDIR) $(CIELPATH)

deps: $(CIELPATH) $(SRCDIR)/go.mod $(SRCDIR)/go.sum
	go mod vendor

$(DISTDIR)/bin/ciel: deps
	export CC
	export CXX
	go build -o $@ -ldflags $(LD_FLAGS) $(REPO_PATH)/cmd/ciel

plugin: plugin/*
	cp -fR $^ $(DISTDIR)/libexec/ciel-plugin

build: $(DISTDIR)/bin/ciel plugin

clean:
	rm -rf $(GOPATH)
	rm -rf $(DISTDIR)
	rm -rf $(SRCDIR)/vendor
	git clean -f -d $(SRCDIR)

install:
	mkdir -p $(DESTDIR)/$(PREFIX)
	cp -R $(DISTDIR)/* $(DESTDIR)/$(PREFIX)

.PHONY: all deps build plugin install clean
