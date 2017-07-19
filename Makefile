PREFIX:=/usr
PLGDIR:=$(PREFIX)/libexec/ciel-plugin
BINDIR:=$(PREFIX)/bin

all:
	git submodule update --init
	sed -e "s|@VERSION@|r`git rev-list --count HEAD`.`git rev-parse --short HEAD`|g" src/ciel/version.go.in > src/ciel/version.go
	GOPATH="$$PWD" go build ciel

install:
	install -Dm755 ciel $(BINDIR)
	mkdir -pm755 $(PLGDIR)
	install -Dm755 ./plugin/* $(PLGDIR)

clean:
	rm ciel
