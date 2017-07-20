PREFIX:=/usr
PLGDIR:=$(PREFIX)/libexec/ciel-plugin
BINDIR:=$(PREFIX)/bin
ARCHS:=amd64 386 arm64 arm mips64le mipsle ppc64 # no powerpc 32 yet

all: ciel

src/ciel/version.go:
	sed -e "s|@VERSION@|r`git rev-list --count HEAD`.`git rev-parse --short HEAD`|g" src/ciel/version.go.in > src/ciel/version.go

ciel: src/ciel/version.go
	GOPATH="$$PWD" go build ciel

test-cross: src/ciel/version.go
	for arch in $(ARCHS); \
	do \
		echo "Cross compiling ciel for Linux/$$arch ..."; \
		GOPATH="$$PWD" GOOS="linux" GOARCH="$$arch" go build -o /tmp/ciel-$$arch ciel; \
	done

test: ciel
	GOPATH="$$PWD" go test ciel
	GOPATH="$$PWD" go test ciel-driver

install: ciel
	install -Dm755 ciel $(BINDIR)
	mkdir -pm755 $(PLGDIR)
	install -Dm755 ./plugin/* $(PLGDIR)

clean:
	rm ciel
	rm src/ciel/version.go

.PHONY: all test-cross test install clean
