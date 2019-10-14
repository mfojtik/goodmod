build:
	mkdir -p ./bin
	GO111MODULE=on go build -o ./bin/gomod-helpers

vendor:
	GO111MODULE=on go mod tidy -v
	GO111MODULE=on go mod vendor -v
