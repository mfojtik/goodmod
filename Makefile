all: build install

build:
	mkdir -p ./bin
	GO111MODULE=on go build -o ./bin/goodmod

build-cross:
	rm -f ./bin/goodmod-linux.gz ./bin/goodmod-darwin.gz
	GOOS=linux GOARCH=amd64 go build -o ./bin/goodmod-linux && gzip ./bin/goodmod-linux
	GOOS=darwin GOARCH=amd64 go build -o ./bin/goodmod-darwin && gzip ./bin/goodmod-darwin

vendor:
	GO111MODULE=on go mod tidy -v
	GO111MODULE=on go mod vendor -v

install:
	cp -f ./bin/goodmod ${GOPATH}/bin/

