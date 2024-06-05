#VERSION = "v2.5.2"
TAG := $(shell git rev-parse --short master)

.PHONY: all

all: linux windows darwin darwin-arm64

linux:
	GOOS=linux GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd64-linux-$(TAG)  \
		  --with github.com/chukmunnlee/caddy-openapi=.

windows:
	GOOS=windows GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd64-windows-$(TAG).exe  \
		  --with github.com/chukmunnlee/caddy-openap=.

darwin:
	GOOS=darwin GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd64-darwin-$(TAG)  \
		  --with github.com/chukmunnlee/caddy-openapi=.

darwin-arm64:
	GOOS=darwin GOARCH=arm64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd64-darwin-$(TAG)  \
		  --with github.com/chukmunnlee/caddy-openapi=.


