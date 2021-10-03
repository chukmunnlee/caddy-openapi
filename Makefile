#VERSION = v2.2.1
VERSION = ""
TAG := $(shell git rev-parse --short master)

.PHONY: all

all: linux windows darwin

linux:
	GOOS=linux GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd6-linux-$(TAG)  \
		  --with github.com/chukmunnlee/caddy-openapi 

windows:
	GOOS=windows GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd6-windows-$(TAG).exe  \
		  --with github.com/chukmunnlee/caddy-openapi

darwin:
	GOOS=darwin GOARCH=amd64 xcaddy build $(VERSION) \
		  --output dist/caddy-amd6-darwin-$(TAG)  \
		  --with github.com/chukmunnlee/caddy-openapi


