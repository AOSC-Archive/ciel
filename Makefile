PREFIX:=/usr/local
CC:=/usr/bin/cc
CXX:=/usr/bin/c++

VERSION=$(shell git describe --tags)
SRCDIR=$(shell pwd)

export GOPATH=$(SRCDIR)/workdir
export CC
export CXX

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

deps: $(CIELPATH)
	go get -u github.com/godbus/dbus

config:
	cp $(SRCDIR)/_config.go $(SRCDIR)/config.go
	sed 's,__VERSION__,$(VERSION),g' -i $(SRCDIR)/config.go
	sed 's,__PREFIX__,$(PREFIX),g' -i $(SRCDIR)/config.go

$(DISTDIR)/bin/ciel: deps config
	export CC
	export CXX
	go build -o $@ ciel

plugin: plugin/*
	cp -fR $^ $(DISTDIR)/libexec/ciel-plugin

build: $(DISTDIR)/bin/ciel plugin

clean:
	rm -rf $(GOPATH)
	rm -rf $(DISTDIR)
	rm -rf $(SRCDIR)/vendor
	git clean -f -d $(SRCDIR)

install:
	mkdir -p $(PREFIX)
	cp -R $(DISTDIR)/* $(PREFIX)

.PHONY: all deps config build plugin install clean
