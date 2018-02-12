ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'vendor' )
endif

all:
	echo "Nothing yet"

verify: vet test lint

test:
	go test $(PKGS)

lint:
	go get -u github.com/golang/lint/golint
	golint -set_exit_status $(PKGS)

vet:
	go vet $(PKGS)
