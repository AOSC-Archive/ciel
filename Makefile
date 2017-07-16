PREFIX:=/usr
PLGDIR:=$(PREFIX)/libexec/ciel-plugin
BINDIR:=$(PREFIX)/bin

all:
	GOPATH="$$PWD" go build ciel

install:
	install -Dm755 ciel $(BINDIR)
	mkdir -pm755 $(PLGDIR)
	install -Dm755 ./plugin/* $(PLGDIR)

clean:
	rm ciel
