PREFIX:=/usr/local

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
	curl -# https://glide.sh/get | sh

deps: $(CIELPATH) $(GLIDE) $(SRCDIR)/glide.yaml
	cd $(CIELPATH)
	$(GLIDE) install
	cd $(SRCDIR)

$(SRCDIR)/config.go: $(SRCDIR)/_config.go
	cp $< $@
	sed 's,__VERSION__,$(VERSION),g' -i $@
	sed 's,__PREFIX__,$(PREFIX),g' -i $@

$(DISTDIR)/bin/ciel: deps $(SRCDIR)/config.go
	go build -o $@ ciel

$(DISTDIR)/libexec/ciel-plugin: plugin/*
	cp -fR $^ $@

build: $(DISTDIR)/bin/ciel $(DISTDIR)/libexec/ciel-plugin

clean:
	-rm -r $(GOPATH)
	-rm -r $(DISTDIR)
	-rm -r $(SRCDIR)/vendor
	git clean -f -d $(SRCDIR)

install:
	cp -R $(DISTDIR)/* $(PREFIX)

.PHONY: all deps build install clean
