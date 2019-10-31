all: build install

build:
	mkdir -p ./bin
	GO111MODULE=on go build -o ./bin/gomod-helpers

build-cross:
	rm -f ./bin/gomod-helpers-linux.gz ./bin/gomod-helpers-darwin.gz
	GOOS=linux GOARCH=amd64 go build -o ./bin/gomod-helpers-linux && gzip ./bin/gomod-helpers-linux
	GOOS=darwin GOARCH=amd64 go build -o ./bin/gomod-helpers-darwin && gzip ./bin/gomod-helpers-darwin

vendor:
	GO111MODULE=on go mod tidy -v
	GO111MODULE=on go mod vendor -v

install:
	cp -f ./bin/gomod-helpers ${GOPATH}/bin/

