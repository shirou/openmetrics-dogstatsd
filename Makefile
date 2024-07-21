VERSION := $(shell git log -1 --date=short --format="%ad-%H" | head -c 19)
BINARYNAME := openmetrics-dogstatsd

build:
	mkdir -p dist
	@go build -o dist -ldflags="-s -w" . 

build-pi:
	mkdir -p dist
	GOOS=linux GOARCH=arm go build -o dist -ldflags="-s -w" .
	upx dist/$(BINARYNAME)
