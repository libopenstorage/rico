ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'vendor' )
endif

all:
	echo "Nothing yet"

verify: vet test lint

test:
	go test $(PKGS)

lint:
	if [ ! -x $(GOPATH)/bin/golint ] ; then go get -u github.com/golang/lint/golint; fi
	golint -set_exit_status $(PKGS)

vet:
	go vet $(PKGS)
